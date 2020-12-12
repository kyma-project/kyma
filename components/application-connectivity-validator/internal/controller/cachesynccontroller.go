package controller

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions/applicationconnector/v1alpha1"
	listers "github.com/kyma-project/kyma/components/application-operator/pkg/client/listers/applicationconnector/v1alpha1"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Application is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Application fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Application"
	// MessageResourceSynced is the message used for an Event fired when a Application
	// is synced successfully
	MessageResourceSynced = "Application resource synced successfully"
)

// Controller is the controller implementation for Application resources
type Controller struct {
	// clientset is a clientset for our own API group
	clientset clientset.Interface

	applicationLister listers.ApplicationLister
	applicationSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// Name of the application CR the controller syncs the cache for
	appName string

	// Cache to store Application CRs
	appCache *gocache.Cache
}

func NewController(
	clientset clientset.Interface,
	applicationInformer informers.ApplicationInformer,
	appName string,
	appCache *gocache.Cache) *Controller {

	controller := &Controller{
		clientset:         clientset,
		applicationLister: applicationInformer.Lister(),
		applicationSynced: applicationInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Applications"),
		appName:           appName,
		appCache:          appCache,
	}

	applicationInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.enqueueApplication,
			UpdateFunc: func(old, new interface{}) {
				controller.enqueueApplication((new))
			},
			DeleteFunc: controller.enqueueApplication,
		})

	return controller
}

// Queues a given application CR if the validator instance belongs to it
func (c *Controller) enqueueApplication(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	if key == c.appName {
		c.workqueue.Add(key)
	}
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Info("Starting Application Cache controller")

	// Wait for the caches to be synced before starting workers
	log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.applicationSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("Starting workers")
	// Launch two workers to process Application resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Info("Started workers")
	<-stopCh
	log.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Application resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to converge the two.
func (c *Controller) syncHandler(key string) error {
	// Get the Application resource with this namespace/name
	application, err := c.applicationLister.Get(key)
	if err != nil {
		// the Application resource may no longer exist, in which case we should delete it from cache
		if errors.IsNotFound(err) {
			c.appCache.Delete(key)
			log.Infof("Deleted the application '%s' from the cache", key)
			return nil
		}

		return err
	}

	//Add or update the cache for Application resource
	applicationClientIDs := c.getClientIDsFromResource(application)
	c.appCache.Set(application.Name, applicationClientIDs, gocache.NoExpiration)
	log.Infof("Added/Updated the application '%s' in the cache", key)
	return nil
}

func (c *Controller) getClientIDsFromResource(application *v1alpha1.Application) []string {
	if application.Spec.CompassMetadata == nil {
		return []string{}
	}

	return application.Spec.CompassMetadata.Authentication.ClientIds
}
