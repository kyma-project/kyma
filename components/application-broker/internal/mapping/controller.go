package mapping

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// maxApplicationMappingProcessRetries is the number of times a application mapping CR will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms
	maxApplicationMappingProcessRetries = 15
)

//go:generate mockery -name=appGetter -output=automock -outpkg=automock -case=underscore
type appGetter interface {
	Get(internal.ApplicationName) (*internal.Application, error)
}

//go:generate mockery -name=nsPatcher -output=automock -outpkg=automock -case=underscore
type nsPatcher interface {
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *corev1.Namespace, err error)
}

// nsBrokerFacade is responsible for managing namespaced ServiceBrokers and creating proper k8s Services for them in the system namespace
//go:generate mockery -name=nsBrokerFacade -output=automock -outpkg=automock -case=underscore
type nsBrokerFacade interface {
	Create(destinationNs string) error
	Delete(destinationNs string) error
	Exist(destinationNs string) (bool, error)
}

//go:generate mockery -name=mappingLister -output=automock -outpkg=automock -case=underscore
type mappingLister interface {
	ListApplicationMappings(application string) ([]*v1alpha1.ApplicationMapping, error)
}

//go:generate mockery -name=nsBrokerSyncer -output=automock -outpkg=automock -case=underscore
type nsBrokerSyncer interface {
	SyncBroker(name string) error
}

// Controller populates local storage with all ApplicationMapping custom resources created in k8s cluster.
type Controller struct {
	queue          workqueue.RateLimitingInterface
	emInformer     cache.SharedIndexInformer
	nsInformer     cache.SharedIndexInformer
	nsPatcher      nsPatcher
	appGetter      appGetter
	nsBrokerFacade nsBrokerFacade
	nsBrokerSyncer nsBrokerSyncer
	mappingSvc     mappingLister
	log            logrus.FieldLogger
}

// New creates new application mapping controller
func New(emInformer cache.SharedIndexInformer, nsInformer cache.SharedIndexInformer, nsPatcher nsPatcher, appGetter appGetter, nsBrokerFacade nsBrokerFacade, nsBrokerSyncer nsBrokerSyncer, log logrus.FieldLogger) *Controller {
	c := &Controller{
		log:            log.WithField("service", "labeler:controller"),
		emInformer:     emInformer,
		nsInformer:     nsInformer,
		queue:          workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		nsPatcher:      nsPatcher,
		appGetter:      appGetter,
		nsBrokerFacade: nsBrokerFacade,
		nsBrokerSyncer: nsBrokerSyncer,
		mappingSvc:     newMappingService(emInformer),
	}

	// EventHandler reacts every time when we add, update or delete ApplicationMapping
	emInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addEM,
		UpdateFunc: c.updateEM,
		DeleteFunc: c.deleteEM,
	})
	return c
}

func (c *Controller) addEM(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling adding event: while adding new application mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) updateEM(old, cur interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err != nil {
		c.log.Errorf("while handling update event: while adding new application mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) deleteEM(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling deletion event: while adding new application mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

// Run starts the controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.shutdownQueueOnStop(stopCh)

	c.log.Info("Starting Application Mappings controller")
	defer c.log.Infof("Shutting down Application Mappings controller")

	if !cache.WaitForCacheSync(stopCh, c.emInformer.HasSynced) {
		c.log.Error("Timeout occurred on waiting for EM informer caches to sync. Shutdown the controller.")
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.nsInformer.HasSynced) {
		c.log.Error("Timeout occurred on waiting for NS informer caches to sync. Shutdown the controller.")
		return
	}

	c.log.Info("EM controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *Controller) shutdownQueueOnStop(stopCh <-chan struct{}) {
	<-stopCh
	c.queue.ShutDown()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	defer c.queue.Done(key)

	err := c.processItem(key.(string))
	switch {
	case err == nil:
		c.queue.Forget(key)

	case c.queue.NumRequeues(key) < maxApplicationMappingProcessRetries:
		c.log.Errorf("Error processing %q (will retry): %v", key, err)
		c.queue.AddRateLimited(key)

	default: // err != nil and err != temporary and too many retries
		c.log.Errorf("Error processing %q (giving up): %v", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	// TODO: In prometheus-operator they use exists to check if we should delete resources, see:
	// https://github.com/coreos/prometheus-operator/blob/master/pkg/alertmanager/operator.go#L364
	// but in k8s they use Lister to check if event should be delete, see:
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/service/service_controller.go#L725
	// We need to check the guarantees of such solutions and choose the best one.
	emObj, emExist, err := c.emInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return errors.Wrapf(err, "while getting object with key %q from the store", key)
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrapf(err, "while getting name and namespace from key %q", key)
	}

	appNs, err := c.getNamespace(namespace)
	if err != nil {
		return err
	}

	if !emExist {
		if err = c.ensureNsNotLabelled(appNs, name); err != nil {
			return err
		}
		return c.ensureNsBrokerNotRegisteredIfNoMappingsOrSync(namespace)
	}
	if err = c.ensureNsLabelled(name, appNs); err != nil {
		return err
	}
	envMapping, ok := emObj.(*v1alpha1.ApplicationMapping)
	if !ok {
		return fmt.Errorf("cannot cast received object to v1alpha1.ApplicationMapping type, type was [%T]", emObj)
	}
	return c.ensureNsBrokerRegisteredAndSynced(envMapping)
}

func (c *Controller) getNamespace(namespace string) (*corev1.Namespace, error) {
	nsObj, nsExist, nsErr := c.nsInformer.GetIndexer().GetByKey(namespace)
	if nsErr != nil {
		return nil, errors.Wrapf(nsErr, "cannot get the namespace: %q", namespace)
	}

	if !nsExist {
		return nil, fmt.Errorf("namespace [%s] not found", namespace)
	}

	reNs, ok := nsObj.(*corev1.Namespace)
	if !ok {
		return nil, fmt.Errorf("cannot cast received object to corev1.Namespace type, type was [%T]", nsObj)
	}
	return reNs, nil
}

func (c *Controller) ensureNsBrokerRegisteredAndSynced(envMapping *v1alpha1.ApplicationMapping) error {
	brokerExist, err := c.nsBrokerFacade.Exist(envMapping.Namespace)
	if err != nil {
		return errors.Wrapf(err, "while checking if namespaced broker exist in namespace [%s]", envMapping.Namespace)
	}
	if brokerExist {
		if err = c.nsBrokerSyncer.SyncBroker(envMapping.Namespace); err != nil {
			return errors.Wrapf(err, "while syncing namespaced broker from namespace [%s]", envMapping.Namespace)
		}
		return nil
	}

	if envMapping.Name != broker.LivenessApplicationSampleName {
		if err = c.nsBrokerFacade.Create(envMapping.Namespace); err != nil {
			return errors.Wrapf(err, "while creating namespaced broker in namespace [%s]", envMapping.Namespace)
		}
	}

	return nil
}

func (c *Controller) ensureNsBrokerNotRegisteredIfNoMappingsOrSync(namespace string) error {
	brokerExist, err := c.nsBrokerFacade.Exist(namespace)
	if err != nil {
		return errors.Wrapf(err, "while checking if namespaced broker exist in namespace [%s]", namespace)
	}
	if !brokerExist {
		return nil
	}
	mappings, err := c.mappingSvc.ListApplicationMappings(namespace)
	if err != nil {
		return errors.Wrapf(err, "while listing application mappings from namespace [%s]", namespace)
	}
	// delete broker only if there'app no application mappings left in the namespace
	if len(mappings) > 0 {
		if err = c.nsBrokerSyncer.SyncBroker(namespace); err != nil {
			return errors.Wrapf(err, "while syncing namespaced broker from namespace [%s]", namespace)
		}
		return nil
	}
	if err = c.nsBrokerFacade.Delete(namespace); err != nil {
		return errors.Wrapf(err, "while removing namespaced broker from namespace [%s]", namespace)
	}
	return nil
}

func (c *Controller) ensureNsNotLabelled(ns *corev1.Namespace, mName string) error {
	nsCopy := ns.DeepCopy()
	c.log.Infof("Deleting AccessLabel: %q, from the namespace - %q", nsCopy.Labels["accessLabel"], nsCopy.Name)

	if mName == broker.LivenessApplicationSampleName {
		delete(nsCopy.Labels, broker.LivenessProbeLabelKey)
	}

	delete(nsCopy.Labels, "accessLabel")

	err := c.patchNs(ns, nsCopy)
	if err != nil {
		return fmt.Errorf("failed to delete AccessLabel from the namespace: %q, %v", nsCopy.Name, err)
	}

	return nil
}

func (c *Controller) ensureNsLabelled(appName string, appNs *corev1.Namespace) error {
	var label string
	label, err := c.getAppAccLabel(appName)
	if err != nil {
		return errors.Wrapf(err, "cannot get AccessLabel from Application: %q", appName)
	}

	err = c.applyNsAccLabel(appNs, label)
	if err != nil {
		return errors.Wrapf(err, "cannot apply AccessLabel to the namespace: %q", appNs.Name)
	}
	return nil
}

func (c *Controller) applyNsAccLabel(ns *corev1.Namespace, label string) error {
	nsCopy := ns.DeepCopy()
	if nsCopy.Labels == nil {
		nsCopy.Labels = make(map[string]string)
	}

	if label != broker.LivenessAccessLabel {
		nsCopy.Labels["accessLabel"] = label
	} else {
		nsCopy.Labels[broker.LivenessProbeLabelKey] = label
	}

	c.log.Infof("Applying AccessLabel: %q to namespace - %q", label, nsCopy.Name)

	err := c.patchNs(ns, nsCopy)
	if err != nil {
		return fmt.Errorf("failed to apply AccessLabel: %q to the namespace: %q, %v", label, nsCopy.Name, err)
	}

	return nil
}

func (c *Controller) patchNs(nsOrig, nsMod *corev1.Namespace) error {
	oldData, err := json.Marshal(nsOrig)
	if err != nil {
		return errors.Wrapf(err, "while marshalling original namespace")
	}
	newData, err2 := json.Marshal(nsMod)
	if err2 != nil {
		return errors.Wrapf(err, "while marshalling modified namespace")
	}

	patch, err3 := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Namespace{})
	if err3 != nil {
		return errors.Wrapf(err, "while creating patch")
	}

	if _, err := c.nsPatcher.Patch(nsMod.Name, types.StrategicMergePatchType, patch); err != nil {
		return fmt.Errorf("failed to patch namespace: %q: %v", nsMod.Name, err)
	}
	return nil
}

func (c *Controller) getAppAccLabel(appName string) (string, error) {
	// get Application from storage
	app, err := c.appGetter.Get(internal.ApplicationName(appName))
	if err != nil {
		return "", errors.Wrapf(err, "while getting application with name: %q", appName)
	}

	if app.AccessLabel == "" {
		return "", fmt.Errorf("Application %q access label is empty", appName)
	}

	return app.AccessLabel, nil
}

func (c *Controller) closeChanOnCtxCancellation(ctx context.Context, ch chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			close(ch)
			return
		}
	}
}
