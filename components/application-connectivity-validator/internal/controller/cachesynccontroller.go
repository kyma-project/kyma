package controller

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/pkg/logger"

	gocache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions/applicationconnector/v1alpha1"
	listers "github.com/kyma-project/kyma/components/application-operator/pkg/client/listers/applicationconnector/v1alpha1"
)

type Controller struct {
	clientset         clientset.Interface
	applicationLister listers.ApplicationLister
	applicationSynced cache.InformerSynced
	workqueue         workqueue.RateLimitingInterface
	appName           string
	appCache          *gocache.Cache
	log               *logger.Logger
}

func NewController(
	log *logger.Logger,
	clientset clientset.Interface,
	applicationInformer informers.ApplicationInformer,
	appName string,
	appCache *gocache.Cache) *Controller {

	controller := &Controller{
		log:               log,
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
		})

	return controller
}

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

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	c.log.WithContext().Info("Starting Application Cache controller")

	c.log.WithContext().Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.applicationSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.log.WithContext().Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	c.log.WithContext().Info("Started workers")
	<-stopCh
	c.log.WithContext().Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)

		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {

	application, err := c.applicationLister.Get(key)
	if err != nil {

		if errors.IsNotFound(err) {
			c.appCache.Delete(key)
			c.log.WithContext().Infof("Deleted the application '%s' from the cache", key)
			return nil
		}

		return err
	}

	applicationClientIDs := c.getClientIDsFromResource(application)
	c.appCache.Set(key, applicationClientIDs, gocache.NoExpiration)
	c.log.WithContext().Infof("Added/Updated the application '%s' in the cache", key)
	return nil
}

func (c *Controller) getClientIDsFromResource(application *v1alpha1.Application) []string {
	if application.Spec.CompassMetadata == nil {
		return []string{}
	}

	return application.Spec.CompassMetadata.Authentication.ClientIds
}
