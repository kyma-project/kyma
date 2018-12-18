package usagekind

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	ukClient "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	sbuInformer "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions/servicecatalog/v1alpha1"
	ukInformer "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions/servicecatalog/v1alpha1"
	ukLister "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/listers/servicecatalog/v1alpha1"
)

const (
	finalizerName = "servicecatalog.kyma.cx/usage-kind-protection"
	indexKind     = "usedByKind"
)

// ProtectionController adds and removes UsageKindProtection finalizer.
type ProtectionController struct {
	queue workqueue.RateLimitingInterface

	bindingUsageInformer     cache.SharedIndexInformer
	bindingUsageListerSynced cache.InformerSynced

	usageKindLister       ukLister.UsageKindLister
	usageKindListerSynced cache.InformerSynced
	ukClient              ukClient.ServicecatalogV1alpha1Interface

	log        logrus.FieldLogger
	maxRetires int

	// test hooks used only in unit tests
	testHookAddFinalizerDone    func()
	testHookProcessDeletionDone func()
}

type namespaceNameKey struct {
	Namespace string
	Name      string
}

// NewProtectionController creates a new instance of ProtectionController.
func NewProtectionController(
	kindInformer ukInformer.UsageKindInformer,
	sbuInformer sbuInformer.ServiceBindingUsageInformer,
	usageKindInterface ukClient.ServicecatalogV1alpha1Interface,
	log logrus.FieldLogger,
) (*ProtectionController, error) {
	serviceBindingUsageInformer := sbuInformer.Informer()

	err := serviceBindingUsageInformer.AddIndexers(cache.Indexers{
		indexKind: func(obj interface{}) ([]string, error) {
			sbu, ok := obj.(*v1alpha1.ServiceBindingUsage)
			if !ok {
				return nil, fmt.Errorf("cannot convert item, expected ServiceBindingUsage, but was %T", obj)
			}
			return []string{sbu.Spec.UsedBy.Kind}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexer for ServiceBindingUsage Informer")
	}

	c := &ProtectionController{
		queue:                    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "UsageKindProtection"),
		bindingUsageInformer:     serviceBindingUsageInformer,
		bindingUsageListerSynced: serviceBindingUsageInformer.HasSynced,
		usageKindLister:          kindInformer.Lister(),
		usageKindListerSynced:    kindInformer.Informer().HasSynced,
		ukClient:                 usageKindInterface,

		log: log.WithField("service", "controller:usage-kind-protection"),

		maxRetires: defaultMaxRetries,
	}

	kindInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.onAddOrUpdateUsageKind,
		UpdateFunc: func(old, new interface{}) {
			c.onAddOrUpdateUsageKind(new)
		},
		// no need to react on deletion
	})

	// TODO: UsageKind must not be deleted before referenced ServiceBindingUsage deletion work is done
	// TODO (implement-sbu-finalizer): after adding finalizer to SBU or other solution, uncomment this
	//sbuInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	DeleteFunc: c.OnDeleteSBU,
	//})
	return c, nil
}

func (c *ProtectionController) onAddOrUpdateUsageKind(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling addition event: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

// OnDeleteSBU reacts on ServiceBindingUsage deletion
func (c *ProtectionController) OnDeleteSBU(event *controller.SBUDeletedEvent) {
	// TODO (implement-sbu-finalizer): this is temporary solution, the method should be removed when SBU finalizer is implemented

	// The processing is not going through the working queue because ServiceBindingUsage cache may be not up to date.
	// Warning: processing without working queue may cause race condition in a corner case - processing deletion in 2 SBU in the same time.
	// Start processing in a separate goroutine to not block the origin process (the main binding usage controller)
	go func() {
		c.log.Debugf("OnDeleteSBU %s/%s", event.Namespace, event.Name)

		uk, err := c.usageKindLister.Get(event.UsedByKind)
		switch {
		case err == nil:
		case apiErrors.IsNotFound(err):
			c.log.Debugf("UsageKind %s not found, ignoring", event.UsedByKind)
			return
		default:
			c.log.Debugf("Could not process SBU deletion %v, will retry. Error: ", event, err.Error())
			c.queue.AddRateLimited(event.UsedByKind)
			return
		}

		if c.isDeletionCandidate(uk) {
			err := c.removeFinalizerIfNotUsed(uk, namespaceNameKey{
				Namespace: event.Namespace,
				Name:      event.Name,
			})

			if err != nil {
				c.log.Debugf("Could not remove finalizer in UsageKind %s, will retry. Error: ", event.UsedByKind, err.Error())
				c.queue.AddRateLimited(event.UsedByKind)
				return
			}
		}
	}()
}

// TODO (implement-sbu-finalizer): the method will be used after implementing SBU finalizer.
func (c *ProtectionController) onDeleteSBU(obj interface{}) {
	sbu, ok := obj.(*v1alpha1.ServiceBindingUsage)
	if !ok {
		c.log.Errorf("incorrect type of incoming object, expected *ServiceBindingUsage but was %T", obj)
		return
	}
	c.queue.Add(sbu.Spec.UsedBy.Kind)
}

// Run begins watching and syncing.
func (c *ProtectionController) Run(stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		c.queue.ShutDown()
	}()

	c.log.Infof("Starting usage kind protection controller")
	defer c.log.Infof("Shutting down usage kind protection controller")

	if !cache.WaitForCacheSync(stopCh, c.bindingUsageListerSynced, c.usageKindListerSynced) {
		c.log.Error("Timeout occurred on waiting for caches to sync. Shutdown the controller.")
		return
	}

	wait.Until(c.worker, time.Second, stopCh)
}

func (c *ProtectionController) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *ProtectionController) processNextWorkItem() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	err := c.processUsageKind(key.(string))
	switch {
	case err == nil:
		c.queue.Forget(key)

	case c.queue.NumRequeues(key) < c.maxRetires:
		c.log.Debugf("Error processing %q (will retry - it's %d of %d): %v", key, c.queue.NumRequeues(key), c.maxRetires, err)
		c.queue.AddRateLimited(key)

	default: // err != nil and too many retries
		c.log.Errorf("Error processing %q (giving up - to many retires): %v", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *ProtectionController) processUsageKind(name string) error {
	uk, err := c.usageKindLister.Get(name)
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		// absence in store means watcher caught the deletion
		c.log.Debugf("UsageKind %s not found, ignoring", name)
		return nil
	default:
		return errors.Wrapf(err, "while getting UsageKind %q", name)
	}

	if c.isDeletionCandidate(uk) {
		return c.removeFinalizerIfNotUsed(uk, namespaceNameKey{})
	}

	if c.needToAddFinalizer(uk) {
		return c.addFinalizer(uk)
	}
	return nil
}

func (c *ProtectionController) needToAddFinalizer(uk *v1alpha1.UsageKind) bool {
	return uk.ObjectMeta.DeletionTimestamp == nil && !containsString(uk.ObjectMeta.Finalizers, finalizerName)
}

func (c *ProtectionController) addFinalizer(uk *v1alpha1.UsageKind) error {
	if c.testHookAddFinalizerDone != nil {
		defer c.testHookAddFinalizerDone()
	}

	ukCopy := uk.DeepCopy()
	ukCopy.ObjectMeta.Finalizers = append(ukCopy.ObjectMeta.Finalizers, finalizerName)
	_, err := c.ukClient.UsageKinds().Update(ukCopy)
	if err != nil {
		return errors.Wrapf(err, "while adding finalizer to UsageKind %s", uk.Name)
	}
	c.log.Debugf("Added protection finalizer to UK %s", uk.Name)
	return nil
}

func (c *ProtectionController) isDeletionCandidate(uk *v1alpha1.UsageKind) bool {
	return uk.ObjectMeta.DeletionTimestamp != nil && containsString(uk.ObjectMeta.Finalizers, finalizerName)
}

func (c *ProtectionController) removeFinalizerIfNotUsed(uk *v1alpha1.UsageKind, key namespaceNameKey) error {
	if c.testHookProcessDeletionDone != nil {
		defer c.testHookProcessDeletionDone()
	}

	isUsed, err := c.isBeingUsedBySBU(uk, key)
	if err != nil {
		return errors.Wrap(err, "while checking if UsageKind is used by any SBU")
	}
	if !isUsed {
		return c.removeFinalizer(uk)
	}
	return nil
}

//TODO (implement-sbu-finalizer): after implementing sbu finalizer keyToSkip arg is not needed
func (c *ProtectionController) isBeingUsedBySBU(uk *v1alpha1.UsageKind, keyToSkip namespaceNameKey) (bool, error) {
	sbuList, err := c.bindingUsageInformer.GetIndexer().ByIndex(indexKind, uk.Name)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Service Binding Usages for UsageKind %s", uk.Name)
	}

	// TODO (implement-sbu-finalizer): filtering should be removed when SBU has finalizer
	filtered := make([]interface{}, 0)
	for _, obj := range sbuList {
		sbu, ok := obj.(*v1alpha1.ServiceBindingUsage)
		if !ok {
			continue
		}
		key := namespaceNameKey{Namespace: sbu.Namespace, Name: sbu.Name}
		if keyToSkip != key {
			filtered = append(filtered, sbu)
		}
	}
	return len(filtered) > 0, nil
}

func (c *ProtectionController) removeFinalizer(uk *v1alpha1.UsageKind) error {
	ukCopy := uk.DeepCopy()
	ukCopy.ObjectMeta.Finalizers = removeString(ukCopy.ObjectMeta.Finalizers, finalizerName)
	_, err := c.ukClient.UsageKinds().Update(ukCopy)
	if err != nil {
		return errors.Wrapf(err, "while removing finalizer from UsageKind %s", uk.Name)
	}
	c.log.Debugf("Removed protection finalizer from UsageKind %s", uk.Name)
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	newSlice := make([]string, 0)
	for _, item := range slice {
		if item == s {
			continue
		}
		newSlice = append(newSlice, item)
	}
	if len(newSlice) == 0 {
		// no need to store empty slice
		newSlice = nil
	}
	return newSlice
}
