package usagekind

import (
	ukInformer "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions/servicecatalog/v1alpha1"
	ukLister "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/listers/servicecatalog/v1alpha1"

	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const defaultMaxRetries = 15

//go:generate mockery -name=SupervisorRegistry -output=automock -outpkg=automock -case=underscore

// SupervisorRegistry provides methods for register/unregister controller.KubernetesResourceSupervisor
type SupervisorRegistry interface {
	Register(k controller.Kind, supervisor controller.KubernetesResourceSupervisor) error
	Unregister(k controller.Kind) error
}

// Controller watcher UsageKind resource and reflects UsageKind instances to registered supervisors in SupervisorRegistry
type Controller struct {
	kindLister    ukLister.UsageKindLister
	queue         workqueue.RateLimitingInterface
	listerSynced  cache.InformerSynced
	maxRetries    int
	kindContainer SupervisorRegistry
	clientPool    dynamic.ClientPool

	log logrus.FieldLogger

	// testHookAsyncOpDone used only in unit tests
	testHookAsyncOpDone func()
}

// NewKindController creates new Controller instance
func NewKindController(
	kindInformer ukInformer.UsageKindInformer,
	kindContainer SupervisorRegistry,
	dynamicClientPool dynamic.ClientPool,
	log logrus.FieldLogger) *Controller {

	c := &Controller{
		kindLister:    kindInformer.Lister(),
		listerSynced:  kindInformer.Informer().HasSynced,
		maxRetries:    defaultMaxRetries,
		kindContainer: kindContainer,
		clientPool:    dynamicClientPool,
		log:           log.WithField("service", "controller:usage-kind"),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "UsageKind"),
	}

	kindInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	})

	return c
}

func (c *Controller) onAdd(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling addition event: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) onUpdate(old, new interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(old)
	if err != nil {
		c.log.Errorf("while handling addition event: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) onDelete(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling deletion event: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

// Run begins watching and syncing.
func (c *Controller) Run(stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		c.queue.ShutDown()
	}()

	c.log.Infof("Starting usage kind controller")
	defer c.log.Infof("Shutting down usage kind controller")

	if !cache.WaitForCacheSync(stopCh, c.listerSynced) {
		c.log.Error("Timeout occurred on waiting for caches to sync. Shutdown the controller.")
		return
	}

	wait.Until(c.worker, time.Second, stopCh)
}

func (c *Controller) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	if c.testHookAsyncOpDone != nil {
		defer c.testHookAsyncOpDone()
	}

	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	err := c.reconcile(key.(string))
	switch {
	case err == nil:
		c.queue.Forget(key)

	case c.queue.NumRequeues(key) < c.maxRetries:
		c.log.Debugf("Error processing %q (will retry - it's %d of %d): %v", key, c.queue.NumRequeues(key), c.maxRetries, err)
		c.queue.AddRateLimited(key)

	default: // err != nil and too many retries
		c.log.Errorf("Error processing %q (giving up - to many retires): %v", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *Controller) reconcile(name string) error {
	usageKind, err := c.kindLister.Get(name)

	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		// absence in store means watcher caught the deletion
		c.log.Debugf("Starting deletion process of UsageKind %q", name)
		if err := c.reconcileDelete(name); err != nil {
			return errors.Wrapf(err, "while deleting UsageKind %q", name)
		}
		c.log.Debugf("UsageKind %q was successfully deleted", name)
		return nil
	default:
		return errors.Wrapf(err, "while getting UsageKind %q", name)
	}

	c.log.Debugf("Starting reconcile UsageKind add process of %s", name)
	defer c.log.Debugf("Reconcile UsageKind add process of %s completed", name)

	if err := c.reconcileAddUpdate(usageKind); err != nil {
		return errors.Wrapf(err, "while processing %s", name)
	}

	return nil
}

func (c *Controller) reconcileDelete(name string) error {
	c.log.Debugf("Unregistering supervisor for usage kind %s", name)
	return c.kindContainer.Unregister(controller.Kind(name))
}

func (c *Controller) reconcileAddUpdate(kind *api.UsageKind) error {
	c.log.Debugf("Registering supervisor for usage kind %s", kind.Name)
	rip, err := newResourceInterfaceProvider(c.clientPool, schema.GroupVersionKind{
		Version: kind.Spec.Resource.Version,
		Group:   kind.Spec.Resource.Group,
		Kind:    kind.Spec.Resource.Kind,
	})
	if err != nil {
		return errors.Wrapf(err, "while creating resource interface provider")
	}

	supervisor := controller.NewGenericSupervisor(rip,
		newLabelManipulator(kind.Spec.LabelsPath),
		c.log.WithField("service", "generic-supervisor").WithField("kind", kind.Name))

	return c.kindContainer.Register(controller.Kind(kind.Name), supervisor)
}
