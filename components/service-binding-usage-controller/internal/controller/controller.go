package controller

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	scTypes "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scInformer "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1beta1"
	scLister "github.com/kubernetes-sigs/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/metric"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/pretty"
	sbuStatus "github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/status"
	sbuTypes "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	svcatSettings "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/settings/v1alpha1"
	sbuClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	sbuInformer "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions/servicecatalog/v1alpha1"
	sbuLister "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/listers/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// defaultMaxRetries is the number of times a ServiceBindingUsage will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(defaultMaxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	defaultMaxRetries = 15
	// LivenessBUCSample name of ServiceBindingUsage used for liveness probe
	LivenessBUCSample = "informer.liveness.probe.service.binding.usage.name"
)

var podPresetOwnerAnnotationKey = fmt.Sprintf("servicebindingusages.%s/owner-name", sbuTypes.SchemeGroupVersion.Group)

//go:generate mockery -name=podPresetModifier -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=kindsSupervisors -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=bindingLabelsFetcher -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=appliedSpecStorage -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=businessMetric -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=sbuGuard -output=automock -outpkg=automock -case=underscore

type (
	podPresetModifier interface {
		UpsertPodPreset(newPodPreset *svcatSettings.PodPreset) error
		EnsurePodPresetDeleted(name, namespace string) error
	}

	kindsSupervisors interface {
		Get(kind Kind) (KubernetesResourceSupervisor, error)
	}

	bindingLabelsFetcher interface {
		Fetch(svcBinding *scTypes.ServiceBinding) (map[string]string, error)
	}

	appliedSpecStorage interface {
		Get(namespace, name string) (*UsageSpec, bool, error)
		Delete(namespace, name string) error
		Upsert(bUsage *sbuTypes.ServiceBindingUsage, applied bool) error
	}

	prefixGetter interface {
		GetPrefix(bUsage *sbuTypes.ServiceBindingUsage) string
	}

	onDeleteListener interface {
		OnDeleteSBU(event *SBUDeletedEvent)
	}

	businessMetric interface {
		RecordError(controllerName string)
		IncrementQueueLength(controllerName string)
		DecrementQueueLength(controllerName string)
		RecordLatency(controllerName string, reconcileTime time.Duration)
	}

	sbuGuard interface {
		AddBindingUsage(key string)
		RemoveBindingUsage(key string)
	}
)

// ServiceBindingUsageController watches ServiceBindingUsage and injects data to given Deployment/Function
type ServiceBindingUsageController struct {
	appliedSpecStorage       appliedSpecStorage
	bindingUsageClient       sbuClient.ServicecatalogV1alpha1Interface
	bindingUsageLister       sbuLister.ServiceBindingUsageLister
	bindingUsageListerSynced cache.InformerSynced
	bindingLister            scLister.ServiceBindingLister
	bindingListerSynced      cache.InformerSynced
	labelsFetcher            bindingLabelsFetcher
	kindsSupervisors         kindsSupervisors
	podPresetModifier        podPresetModifier
	maxRetires               int
	guard                    sbuGuard
	log                      logrus.FieldLogger
	queue                    workqueue.RateLimitingInterface
	prefixGetter             prefixGetter
	metric                   businessMetric

	// testHookAsyncOpDone used only in unit tests
	testHookAsyncOpDone func()

	onDeleteListeners []onDeleteListener
}

// NewServiceBindingUsage creates a new ServiceBindingUsageController.
func NewServiceBindingUsage(
	appliedSpecStorage appliedSpecStorage,
	bindingUsageClient sbuClient.ServicecatalogV1alpha1Interface,
	sbuInformer sbuInformer.ServiceBindingUsageInformer,
	bindingInformer scInformer.ServiceBindingInformer,
	kindSupervisors kindsSupervisors,
	podPresetModifier podPresetModifier,
	labelsFetcher bindingLabelsFetcher,
	sbuGuard sbuGuard,
	log logrus.FieldLogger,
	cbm businessMetric) *ServiceBindingUsageController {
	c := &ServiceBindingUsageController{
		appliedSpecStorage:       appliedSpecStorage,
		bindingUsageClient:       bindingUsageClient,
		bindingUsageLister:       sbuInformer.Lister(),
		bindingUsageListerSynced: sbuInformer.Informer().HasSynced,
		bindingLister:            bindingInformer.Lister(),
		bindingListerSynced:      bindingInformer.Informer().HasSynced,
		kindsSupervisors:         kindSupervisors,
		podPresetModifier:        podPresetModifier,
		labelsFetcher:            labelsFetcher,
		guard:                    sbuGuard,
		maxRetires:               defaultMaxRetries,
		log:                      log.WithField("service", "controller:service-binding-usage"),
		queue:                    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ServiceBindingUsage"),
		prefixGetter:             &envPrefixGetter{},
		metric:                   cbm,
	}

	bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.triggerServiceBindingUsageReconciliation,
	})

	sbuInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAddServiceBindingUsage,
		UpdateFunc: c.onUpdateOrRelistServiceBindingUsage,
		DeleteFunc: c.onDeleteServiceBindingUsage,
	})

	return c
}

func (c *ServiceBindingUsageController) onAddServiceBindingUsage(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling addition event: couldn't get key: %v", err)
		c.metric.RecordError(metric.SbuController)
		return
	}
	c.log.Infof("new add event with key %q triggered", key)
	c.queue.Add(key)
	c.metric.IncrementQueueLength(metric.SbuController)
}

func (c *ServiceBindingUsageController) onDeleteServiceBindingUsage(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling deletion event: couldn't get key: %v", err)
		c.metric.RecordError(metric.SbuController)
		return
	}
	c.log.Infof("new delete event with key %q triggered", key)
	c.queue.Add(key)
	c.metric.IncrementQueueLength(metric.SbuController)
}

func (c *ServiceBindingUsageController) onUpdateOrRelistServiceBindingUsage(old, cur interface{}) {
	oldUsage, ok := old.(*sbuTypes.ServiceBindingUsage)
	if !ok {
		c.log.Warnf("while handling update: cannot covert obj [%+v] of type %T to *ServiceBindingUsage", cur, cur)
		return
	}

	curUsage, ok := cur.(*sbuTypes.ServiceBindingUsage)
	if !ok {
		c.log.Warnf("while handling update: cannot covert obj [%+v] of type %T to *ServiceBindingUsage", cur, cur)
		return
	}

	if !c.isUpdateNeeded(oldUsage, curUsage) {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err != nil {
		c.log.Errorf("while handling updating event: couldn't get key: %v", err)
		c.metric.RecordError(metric.SbuController)
		return
	}
	c.log.Infof("new update event with key %q triggered. Updated object: %s/%s", key, oldUsage.Namespace, oldUsage.Name)
	c.queue.Add(key)
	c.metric.IncrementQueueLength(metric.SbuController)
}

func (c *ServiceBindingUsageController) triggerServiceBindingUsageReconciliation(_, cur interface{}) {
	sb, ok := cur.(*scTypes.ServiceBinding)
	if !ok {
		c.log.Errorf("Error handling ServiceBinding update: cannot covert obj [%+v] of type %T to *ServiceBindingUsage", cur, cur)
		return
	}
	if !isServiceBindingReady(sb) {
		c.log.Infof("ServiceBinding %s/%s is not ready. ServiceBindingUsage trigger action delayed", sb.Namespace, sb.Name)
		return
	}

	sbus, err := c.bindingUsageLister.ServiceBindingUsages(sb.Namespace).List(labels.NewSelector())
	if err != nil {
		c.log.Errorf("Error listing ServiceBindingUsage in %q namespace: %s", sb.Namespace, err)
		return
	}

	for _, sbu := range sbus {
		if sbu.Spec.ServiceBindingRef.Name != sb.Name {
			continue
		}
		if isServiceBindingUsageReady(sbu) {
			continue
		}

		toUpdate := sbu.DeepCopy()
		toUpdate.Spec.ReprocessRequest = toUpdate.Spec.ReprocessRequest + 1

		_, err = c.bindingUsageClient.ServiceBindingUsages(toUpdate.Namespace).Update(toUpdate)
		if err != nil {
			c.log.Errorf("Error updating ServiceBindingUsage %s/%s", toUpdate.Namespace, toUpdate.Name)
			return
		}
		c.log.Infof("ServiceBindingUsage %s/%s triggered", toUpdate.Namespace, toUpdate.Name)
	}
}

// Run begins watching and syncing.
func (c *ServiceBindingUsageController) Run(stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		c.queue.ShutDown()
	}()

	c.log.Infof("Starting service binding usage controller")
	defer c.log.Infof("Shutting down service binding usage controller")

	if !cache.WaitForCacheSync(stopCh, c.bindingUsageListerSynced,
		c.bindingListerSynced) {
		c.log.Error("Timeout occurred on waiting for caches to sync. Shutdown the controller.")
		return
	}

	wait.Until(c.worker, time.Second, stopCh)
}

func (c *ServiceBindingUsageController) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *ServiceBindingUsageController) processNextWorkItem() bool {
	if c.testHookAsyncOpDone != nil {
		defer c.testHookAsyncOpDone()
	}

	key, shutdown := c.queue.Get()
	reconcileStart := time.Now()
	if shutdown {
		c.log.Info("queue has been shutdown")
		return false
	}
	defer func() {
		c.log.Infof("process for key %q has been completed", key)
		c.queue.Done(key)
		c.metric.RecordLatency(metric.SbuController, time.Now().Sub(reconcileStart))
	}()

	namespace, name, err := cache.SplitMetaNamespaceKey(key.(string))
	if err != nil {
		c.log.Errorf("Error processing %q (splitting meta namespace key failed): %v", key, err)
		c.metric.RecordError(metric.SbuController)
		c.queue.Forget(key)
		c.metric.DecrementQueueLength(metric.SbuController)
		return true
	}
	// Skip all reconcile process if ServiceBindingUsage comes from informer liveness probe
	// in that case we only check informer handle the queue so we need to only change the SBU status,
	// all process is not needed
	if name == LivenessBUCSample {
		err := c.handleServiceBindingUsageSample(namespace, name)
		if err != nil {
			c.log.Errorf("failed handle SBU sample: %s", err)
			c.metric.RecordError(metric.SbuController)
		}
		c.queue.Forget(key)
		c.metric.DecrementQueueLength(metric.SbuController)
		return true
	}

	retry := c.queue.NumRequeues(key)
	usageStatus, err := c.syncServiceBindingUsage(namespace, name)
	switch {
	case err == nil:
		c.queue.Forget(key)
		c.metric.DecrementQueueLength(metric.SbuController)

	case retry < c.maxRetires:
		c.log.Debugf("Error processing %q (will retry - it's %d of %d): %v", key, retry, c.maxRetires, err)
		c.queue.AddRateLimited(key)

	default: // err != nil and too many retries
		c.log.Errorf("Error processing %q (giving up - to many retires): %v", key, err)
		c.metric.RecordError(metric.SbuController)
		c.queue.Forget(key)
		c.metric.DecrementQueueLength(metric.SbuController)
	}

	// set ServiceBindingUsage status if ServiceBindingUsage and his status exist
	if usageStatus != nil {
		c.log.Debug("Starting process of updating ServiceBindingUsageCondition")
		usageStatus.wrapMessageForFailed(fmt.Sprintf("Process error during %d attempts from %d", retry, c.maxRetires))

		bindingUsage, err := c.bindingUsageLister.ServiceBindingUsages(namespace).Get(name)
		if err != nil {
			c.log.Errorf("Cannot get ServiceBindingUsage %s/%s, got error: %v", namespace, name, err)
			c.metric.RecordError(metric.SbuController)
			return true
		}

		c.log.Debugf("Updating %q conditions", pretty.ServiceBindingUsageName(bindingUsage))
		condition := sbuStatus.NewUsageCondition(usageStatus.sbuType, usageStatus.condition, usageStatus.reason, usageStatus.message)
		if err := c.updateStatus(bindingUsage, *condition); err != nil {
			c.log.Errorf("Error processing %q while updating sbu status with condition %+v", key, condition)
			c.metric.RecordError(metric.SbuController)
		}
	}

	return true
}

func (c *ServiceBindingUsageController) syncServiceBindingUsage(namespace string, name string) (*bindingUsageStatus, error) {
	// holds the latest ServiceBindingUsage info from apiserver
	bindingUsage, err := c.bindingUsageLister.ServiceBindingUsages(namespace).Get(name)
	bindingUsageStatus := newBindingUsageStatus()

	switch {
	case err == nil:
		bindingUsageStatus.condition = sbuTypes.ConditionTrue
	case apiErrors.IsNotFound(err):
		// absence in store means watcher caught the deletion
		c.log.Debugf("Starting deletion process of ServiceBindingUsage %q", pretty.KeyItem(namespace, name))
		c.guard.RemoveBindingUsage(pretty.Key(namespace, name))
		if err := c.reconcileServiceBindingUsageDelete(namespace, name); err != nil {
			// TODO(adding finalizer): add a status update in case of error
			// in the same way as we have for `reconcileServiceBindingUsageAdd`
			return nil, errors.Wrapf(err, "while deleting ServiceBidingUsage %q", pretty.KeyItem(namespace, name))
		}
		c.log.Debugf("ServiceBindingUsage %q was successfully deleted", pretty.KeyItem(namespace, name))
		return nil, nil
	default:
		return nil, errors.Wrap(err, "while getting ServiceBindingUsage")
	}

	c.log.Debugf("Starting reconcile ServiceBindingUsage add process of %s", pretty.ServiceBindingUsageName(bindingUsage))
	defer c.log.Debugf("Reconcile ServiceBindingUsage add process of %s completed", pretty.KeyItem(namespace, name))

	if err := c.reconcileServiceBindingUsageAdd(bindingUsage); err != nil {
		bindingUsageStatus.condition = sbuTypes.ConditionFalse
		bindingUsageStatus.reason = err.Reason
		bindingUsageStatus.message = err.Message

		return bindingUsageStatus, errors.Wrapf(err, "while processing %s", pretty.ServiceBindingUsageName(bindingUsage))
	}

	c.guard.AddBindingUsage(pretty.Key(namespace, name))
	return bindingUsageStatus, nil
}

func (c *ServiceBindingUsageController) handleServiceBindingUsageSample(namespace, name string) error {
	bindingUsage, err := c.bindingUsageLister.ServiceBindingUsages(namespace).Get(name)

	switch {
	case err == nil:
		if err := c.updateStatus(bindingUsage, sbuTypes.ServiceBindingUsageCondition{
			Status:             sbuTypes.ConditionTrue,
			LastTransitionTime: metaV1.Now(),
		}); err != nil {
			c.log.Errorf("Error processing %q while updating sbu status for SBU sample: %s", err)
		}
	case apiErrors.IsNotFound(err):
		// absence in store means watcher caught the deletion
		return nil
	default:
		return errors.Wrap(err, "while getting ServiceBindingUsage")
	}

	return nil
}

func (c *ServiceBindingUsageController) reconcileServiceBindingUsageAdd(newUsage *sbuTypes.ServiceBindingUsage) *processBindingUsageError {
	c.log.Debugf("process of reconcile %s", pretty.ServiceBindingUsageName(newUsage))
	var (
		workNS         = newUsage.Namespace
		newBindingName = newUsage.Spec.ServiceBindingRef.Name
	)

	svcBinding, err := c.bindingLister.ServiceBindings(workNS).Get(newBindingName)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			if err := c.ensureOwnerRefNotExists(newUsage); err != nil {
				c.log.Errorf("%v: while deleting OwnerReferences from sbu %q", err)
			}
		}
		return newProcessServiceBindingError(
			sbuStatus.ServiceBindingGetErrorReason,
			errors.Wrapf(err, "while getting ServiceBinding %q from namespace %q", newBindingName, workNS),
		)
	}

	if err := c.ensureOwnerRef(newUsage, svcBinding); err != nil {
		return newProcessServiceBindingError(
			sbuStatus.AddOwnerReferenceErrorReason,
			errors.Wrapf(err, "while adding OwnerReference to %s", pretty.ServiceBindingUsageName(newUsage)),
		)
	}

	if svcBinding.Status.AsyncOpInProgress {
		return newProcessServiceBindingError(
			sbuStatus.ServiceBindingOngoingAsyncOptReason,
			fmt.Errorf("cannot use %s which has ongoing asynchronous operation", pretty.ServiceBindingName(svcBinding)),
		)
	}

	if !isServiceBindingReady(svcBinding) {
		return newProcessServiceBindingError(
			sbuStatus.ServiceBindingNotReadyReason,
			fmt.Errorf("cannot use %s which is not in ready state", pretty.ServiceBindingName(svcBinding)),
		)
	}

	newPodPreset := c.createPodPresetForBindingUsage(newUsage)
	// Upsert - thanks to that we always have proper PodPreset in place
	if err := c.podPresetModifier.UpsertPodPreset(newPodPreset); err != nil {
		return newProcessServiceBindingError(
			sbuStatus.PodPresetUpsertErrorReason,
			errors.Wrapf(err, "while upserting the %s", pretty.PodPresetName(newPodPreset)),
		)
	}

	bindingLabels, err := c.labelsFetcher.Fetch(svcBinding)
	if err != nil {
		return newProcessServiceBindingError(
			sbuStatus.FetchBindingLabelsErrorReason,
			errors.Wrapf(err, "while fetching bindings labels for binding [%s]", pretty.ServiceBindingName(svcBinding)),
		)
	}

	labelsToApply, err := Merge(newPodPreset.Spec.Selector.MatchLabels, bindingLabels)
	if err != nil {
		return newProcessServiceBindingError(
			sbuStatus.ApplyLabelsConflictErrorReason,
			errors.Wrapf(err, "while merging labels: from PodPreset selector[%v] with binding labels [%v]", newPodPreset.Spec.Selector.MatchLabels, bindingLabels),
		)
	}

	if err := c.ensureProperKindIsLabeled(newUsage, labelsToApply); err != nil {
		return newProcessServiceBindingError(
			sbuStatus.EnsureLabelsAppliedErrorReason,
			errors.Wrapf(err, "while ensuring proper labels on kind %s", newUsage.Spec.UsedBy.Kind),
		)
	}

	c.log.Debugf("process for %s has been completed", pretty.ServiceBindingUsageName(newUsage))
	return nil
}

func (c *ServiceBindingUsageController) ensureOwnerRef(newUsage *sbuTypes.ServiceBindingUsage, binding *scTypes.ServiceBinding) error {
	for _, ref := range newUsage.OwnerReferences {
		if ref.Kind == "ServiceBinding" && ref.Name == binding.Name {
			return nil
		}
	}

	newUsage.OwnerReferences = append(newUsage.OwnerReferences, metaV1.OwnerReference{
		APIVersion: "servicecatalog.k8s.io/v1beta1",
		Kind:       "ServiceBinding",
		Name:       binding.Name,
		UID:        binding.UID,
	})

	if _, err := c.bindingUsageClient.ServiceBindingUsages(newUsage.Namespace).Update(newUsage); err != nil {
		return errors.Wrapf(err, "while updating %s", pretty.ServiceBindingUsageName(newUsage))
	}

	return nil
}

func (c *ServiceBindingUsageController) ensureOwnerRefNotExists(newUsage *sbuTypes.ServiceBindingUsage) error {
	if len(newUsage.OwnerReferences) == 0 {
		return nil
	}

	ownerReferences := make([]metaV1.OwnerReference, 0)
	for _, ref := range newUsage.OwnerReferences {
		if ref.Kind != "ServiceBinding" {
			ownerReferences = append(ownerReferences, ref)
		}
	}

	newUsage.OwnerReferences = ownerReferences
	if _, err := c.bindingUsageClient.ServiceBindingUsages(newUsage.Namespace).Update(newUsage); err != nil {
		return errors.Wrapf(err, "while updating %s", pretty.ServiceBindingUsageName(newUsage))
	}

	return nil
}

func (c *ServiceBindingUsageController) ensureProperKindIsLabeled(newUsage *sbuTypes.ServiceBindingUsage, labelsToApply map[string]string) error {
	var (
		workNs    = newUsage.Namespace
		usageName = newUsage.Name
	)

	storedSpec, found, err := c.appliedSpecStorage.Get(workNs, usageName)
	if err != nil {
		return errors.Wrapf(err, "while getting stored Spec for %s", pretty.ServiceBindingUsageName(newUsage))
	}

	if !found {
		if err := c.ensureNewLabels(newUsage, labelsToApply); err != nil {
			return errors.Wrap(err, "while applying labels")
		}
		return nil
	}

	appliedLabelsOrigin := labelsOrigin{
		UsedBySpec: storedSpec.UsedBy,
		Namespace:  workNs,
		UsageName:  usageName,
	}

	usedBySpecEqual := c.isUsedBySpecEqual(storedSpec.UsedBy, newUsage.Spec.UsedBy)
	labelsEqual, err := c.isLabelsEqual(appliedLabelsOrigin, labelsToApply)
	switch {
	case err == nil:
	case IsNotFoundError(err) && !usedBySpecEqual:
		// Scenario: someone created SBU with not exiting deployment, then modified SBU to point to the new deploy,
		// so we are receiving event about update and when we checking if labels are equal then the
		// previous deployment still does not exits but `spec` was modified so we need to proceed further
	default:
		return errors.Wrap(err, "while checking if applied labels are equal with current ones")
	}

	if !usedBySpecEqual || !labelsEqual {
		err := c.revertLabels(workNs, usageName, storedSpec.UsedBy)
		if err != nil && !IsNotFoundError(err) {
			return errors.Wrap(err, "while reverting old labels")
		}

		if err := c.ensureNewLabels(newUsage, labelsToApply); err != nil {
			return errors.Wrap(err, "while applying labels")
		}
	}

	if usedBySpecEqual && labelsEqual && !storedSpec.Applied {
		if err := c.ensureNewLabels(newUsage, labelsToApply); err != nil {
			return errors.Wrap(err, "while applying labels")
		}
	}

	return nil
}

func (c *ServiceBindingUsageController) revertLabels(usageNamespace, usageName string, storedUsedBySpec sbuTypes.LocalReferenceByKindAndName) error {
	previousSupervisor, err := c.kindsSupervisors.Get(Kind(storedUsedBySpec.Kind))
	if err != nil {
		return errors.Wrapf(err, "while getting supervisor for kind %q", Kind(storedUsedBySpec.Kind))
	}

	// revert
	if err := previousSupervisor.EnsureLabelsDeleted(usageNamespace, storedUsedBySpec.Name, usageName); err != nil {
		return errors.Wrapf(err, "while trying to revert changes made on %s %s/%s", storedUsedBySpec.Kind, usageNamespace, storedUsedBySpec.Name)
	}

	// changes reverted - delete old spec
	if err := c.appliedSpecStorage.Delete(usageNamespace, usageName); err != nil {
		return errors.Wrap(err, "while deleting from storage the old Spec")
	}

	return nil
}

func (c *ServiceBindingUsageController) ensureNewLabels(newUsage *sbuTypes.ServiceBindingUsage, labelsToApply map[string]string) error {
	currentKindSupervisor, err := c.kindsSupervisors.Get(Kind(newUsage.Spec.UsedBy.Kind))
	if err != nil {
		return errors.Wrapf(err, "while getting concrete supervisor for kind %q", Kind(newUsage.Spec.UsedBy.Kind))
	}

	if err := c.appliedSpecStorage.Upsert(newUsage, false); err != nil {
		return errors.Wrapf(err, "while saving spec for %s", pretty.ServiceBindingUsageName(newUsage))
	}

	if err := currentKindSupervisor.EnsureLabelsCreated(newUsage.Namespace, newUsage.Spec.UsedBy.Name, newUsage.Name, labelsToApply); err != nil {
		return errors.Wrapf(err, "while ensuring labels on %q %q in namespace %q", Kind(newUsage.Spec.UsedBy.Kind), newUsage.Spec.UsedBy.Name, newUsage.Namespace)
	}

	if err := c.appliedSpecStorage.Upsert(newUsage, true); err != nil {
		return errors.Wrapf(err, "while saving spec for %s", pretty.ServiceBindingUsageName(newUsage))
	}
	return nil
}

func (c *ServiceBindingUsageController) isUsedBySpecEqual(specA, specB sbuTypes.LocalReferenceByKindAndName) bool {
	return reflect.DeepEqual(specA, specB)
}

type labelsOrigin struct {
	UsedBySpec sbuTypes.LocalReferenceByKindAndName
	Namespace  string
	UsageName  string
}

func (c *ServiceBindingUsageController) isLabelsEqual(lSource labelsOrigin, labels map[string]string) (bool, error) {
	concreteSupervisor, err := c.kindsSupervisors.Get(Kind(lSource.UsedBySpec.Kind))
	if err != nil {
		return false, errors.Wrapf(err, "while getting concrete supervisor for kind %q", Kind(lSource.UsedBySpec.Kind))
	}

	appliedLabels, err := concreteSupervisor.GetInjectedLabels(lSource.Namespace, lSource.UsedBySpec.Name, lSource.UsageName)
	if err != nil {
		return false, errors.Wrap(err, "while getting injected labels")
	}

	if len(appliedLabels) != len(labels) {
		return false, nil
	}

	for key, originValue := range appliedLabels {
		if toApplyValue, exists := labels[key]; !exists || originValue != toApplyValue {
			return false, nil
		}
	}

	return true, nil
}

func (c *ServiceBindingUsageController) updateStatus(bUsage *sbuTypes.ServiceBindingUsage, condition sbuTypes.ServiceBindingUsageCondition) error {
	copyUsage := bUsage.DeepCopy()
	sbuStatus.SetUsageCondition(&copyUsage.Status, condition)
	_, err := c.bindingUsageClient.ServiceBindingUsages(copyUsage.Namespace).Update(copyUsage)
	if err != nil {
		return errors.Wrapf(err, "while updating status of %s", pretty.ServiceBindingUsageName(copyUsage))
	}

	return nil
}

func (c *ServiceBindingUsageController) isUpdateNeeded(specA *sbuTypes.ServiceBindingUsage, specB *sbuTypes.ServiceBindingUsage) bool {
	return !reflect.DeepEqual(specA.Spec, specB.Spec)
}

func (c *ServiceBindingUsageController) createPodPresetForBindingUsage(bUsage *sbuTypes.ServiceBindingUsage) *svcatSettings.PodPreset {
	return &svcatSettings.PodPreset{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: bUsage.Namespace,
			Name:      c.podPresetNameFromBindingUsageName(bUsage.Name),
			Annotations: map[string]string{
				podPresetOwnerAnnotationKey: bUsage.Name,
			},
		},
		Spec: svcatSettings.PodPresetSpec{
			Selector: metaV1.LabelSelector{
				MatchLabels: c.podPresetMatchLabels(bUsage),
			},
			EnvFrom: []coreV1.EnvFromSource{
				{
					Prefix: c.prefixGetter.GetPrefix(bUsage),
					SecretRef: &coreV1.SecretEnvSource{
						LocalObjectReference: coreV1.LocalObjectReference{
							Name: bUsage.Spec.ServiceBindingRef.Name,
						},
					},
				},
			},
		},
	}
}

func (c *ServiceBindingUsageController) podPresetMatchLabels(bUsage *sbuTypes.ServiceBindingUsage) map[string]string {
	key := fmt.Sprintf("use-%s", bUsage.UID)

	return map[string]string{
		key: bUsage.ResourceVersion,
	}
}

func (c *ServiceBindingUsageController) reconcileServiceBindingUsageDelete(usageNamespace, usageName string) *processBindingUsageError {
	if err := c.podPresetModifier.EnsurePodPresetDeleted(usageNamespace, c.podPresetNameFromBindingUsageName(usageName)); err != nil {
		return newProcessServiceBindingError(
			sbuStatus.PodPresetDeleteErrorReason,
			errors.Wrap(err, "while ensuring that PodPreset is deleted"),
		)
	}

	storedSpec, found, err := c.appliedSpecStorage.Get(usageNamespace, usageName)
	if err != nil {
		return newProcessServiceBindingError(
			sbuStatus.GetStoredSpecError,
			errors.Wrapf(err, "while getting stored Spec for %s/%s", usageNamespace, usageName),
		)
	}

	if !found {
		return nil
	}

	if err := c.revertLabels(usageNamespace, usageName, storedSpec.UsedBy); err != nil {
		return newProcessServiceBindingError(
			sbuStatus.EnsureLabelsDeletedErrorReason,
			errors.Wrap(err, "while reverting old labels"),
		)
	}

	c.informListeners(&SBUDeletedEvent{
		Name:       usageName,
		Namespace:  usageNamespace,
		UsedByKind: storedSpec.UsedBy.Kind,
	})

	return nil
}

func (c *ServiceBindingUsageController) podPresetNameFromBindingUsageName(bindingUsageName string) string {
	h := sha1.New()
	h.Write([]byte(bindingUsageName))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *ServiceBindingUsageController) informListeners(event *SBUDeletedEvent) {
	for _, listener := range c.onDeleteListeners {
		listener.OnDeleteSBU(event)
	}
}

// AddOnDeleteListener adds OnDeleteListener
// The method is not thread safe
func (c *ServiceBindingUsageController) AddOnDeleteListener(listener onDeleteListener) {
	c.onDeleteListeners = append(c.onDeleteListeners, listener)
}

// isServiceBindingUsageReady returns whether the given service binding usage has a ready condition with status true.
func isServiceBindingUsageReady(sbu *sbuTypes.ServiceBindingUsage) bool {
	for _, cond := range sbu.Status.Conditions {
		if cond.Type == sbuTypes.ServiceBindingUsageReady {
			return cond.Status == sbuTypes.ConditionTrue
		}
	}

	return false
}

// isServiceBindingReady returns whether the given service binding has a ready condition
// with status true.
//
// I checked that they always updated this status to false if there are some problems.
// see: https://github.com/kubernetes-sigs/service-catalog/blob/v0.1.3/pkg/controller/controller_binding.go#L178
//
// What's more they doing same thing for checking if given service instance is ready.
// see: https://github.com/kubernetes-sigs/service-catalog/blob/v0.1.3/pkg/controller/controller.go#L606
func isServiceBindingReady(instance *scTypes.ServiceBinding) bool {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == scTypes.ServiceBindingConditionReady {
			return cond.Status == scTypes.ConditionTrue
		}
	}

	return false
}
