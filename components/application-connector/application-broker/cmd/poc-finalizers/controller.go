package main

import (
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	mV1alpha12 "github.com/kyma-project/kyma/components/application-connector/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	cV1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions/applicationconnector/v1alpha1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// FinalizerName is the protection finalizer name
const FinalizerName = "protection-finalizer"

// Controller add finalizers logic the Application resources. Blocks deletion until all connected ApplicationMapping are removed.
type Controller struct {
	queue        workqueue.RateLimitingInterface
	reInformer   cache.SharedIndexInformer
	emInterface  mV1alpha12.ApplicationMappingInterface
	appInterface cV1alpha12.ApplicationInterface
	log          *logrus.Entry
}

// NewProtectionController creates protection controller instance
func NewProtectionController(applicationInformer informers.ApplicationInformer,
	emInterface mV1alpha12.ApplicationMappingInterface,
	environmentInterface cV1alpha12.ApplicationInterface,
	log *logrus.Entry) *Controller {

	reInformer := applicationInformer.Informer()

	c := &Controller{
		queue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reprotection"),
		reInformer:   applicationInformer.Informer(),
		emInterface:  emInterface,
		log:          log.WithField("service", "app-protection-controller"),
		appInterface: environmentInterface,
	}

	reInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.log.Infof("Event: added")
			c.reAddedUpdated(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			c.log.Infof("Event: updated")
			c.reAddedUpdated(new)
		},
		DeleteFunc: func(obj interface{}) {
			c.log.Infof("Event: deleted")
		},
	})

	return c
}

func (c *Controller) reAddedUpdated(obj interface{}) {
	app, ok := obj.(*v1alpha1.Application)
	if !ok {
		c.log.Errorf("Application informer returned non-Application object: %#v", obj)
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(app)
	if err != nil {
		c.log.Errorf("couldn't get key for Application %#v: %v", app, err)
		return
	}
	c.log.Infof("Got Application: %s", key)

	if needToAddFinalizer(app) || isDeletionCandidate(app) {
		c.queue.AddRateLimited(key)
	}
}

// Run runs the controller
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.log.Info("Starting protection controller")
	defer c.log.Info("Shutting down protection controller")

	if !cache.WaitForCacheSync(stopCh) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	c.reInformer.Run(stopCh)

	<-stopCh
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem process item from the queue and returns false when it's time to quit.
func (c *Controller) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.processRE(key.(string))
	if err != nil {
		c.log.Errorf("Could not process Application %s: %s", key, err.Error())
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	return true
}

func (c *Controller) processRE(key string) error {
	_, appName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	app, err := c.appInterface.Get(appName, v1.GetOptions{})
	if err != nil {
		return err
	}

	if needToAddFinalizer(app) {
		clone := app.DeepCopy()
		clone.ObjectMeta.Finalizers = append(clone.ObjectMeta.Finalizers, FinalizerName)
		c.log.Info("Adding finalizer")
		_, err := c.appInterface.Update(clone)
		c.log.Infof("Finalizer added")
		if err != nil {
			return err
		}
	}

	if isDeletionCandidate(app) {
		items, _ := c.emInterface.List(v1.ListOptions{})
		exists := false

		// find if application mapping exists
		for _, item := range items.Items {
			if item.Name == appName {
				exists = true
				break
			}
		}

		// if ApplicationMapping does not exists - remove finalizer
		if !exists {
			clone := app.DeepCopy()
			clone.ObjectMeta.Finalizers = removeString(clone.ObjectMeta.Finalizers, FinalizerName)

			c.log.Info("Removing finalizer")
			_, err := c.appInterface.Update(clone)
			c.log.Infof("Finalizer removed")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isDeletionCandidate(app *v1alpha1.Application) bool {
	return app.ObjectMeta.DeletionTimestamp != nil && containsString(app.ObjectMeta.Finalizers, FinalizerName)
}

func needToAddFinalizer(app *v1alpha1.Application) bool {
	return app.ObjectMeta.DeletionTimestamp == nil && !containsString(app.ObjectMeta.Finalizers, FinalizerName)
}

func removeString(slice []string, s string) []string {
	newSlice := make([]string, 0)
	for _, item := range slice {
		if item == s {
			continue
		}
		newSlice = append(newSlice, item)
	}
	return newSlice
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
