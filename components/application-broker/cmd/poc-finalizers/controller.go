package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	v1alpha12 "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	informers "github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions/applicationconnector/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// FinalizerName is the protection finalizer name
const FinalizerName = "protection-finalizer"

// Controller add finalizers logic the RemoteEnvironment resources. Blocks deletion until all connected EnvironmentMapping are removed.
type Controller struct {
	queue       workqueue.RateLimitingInterface
	reInformer  cache.SharedIndexInformer
	emInterface v1alpha12.EnvironmentMappingInterface
	reInterface v1alpha12.RemoteEnvironmentInterface
	log         *logrus.Entry
}

// NewProtectionController creates protection controller instance
func NewProtectionController(remoteEnvironmentInformer informers.RemoteEnvironmentInformer,
	emInterface v1alpha12.EnvironmentMappingInterface,
	environmentInterface v1alpha12.RemoteEnvironmentInterface,
	log *logrus.Entry) *Controller {

	reInformer := remoteEnvironmentInformer.Informer()

	c := &Controller{
		queue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "reprotection"),
		reInformer:  remoteEnvironmentInformer.Informer(),
		emInterface: emInterface,
		log:         log.WithField("service", "re-protection-controller"),
		reInterface: environmentInterface,
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
	re, ok := obj.(*v1alpha1.RemoteEnvironment)
	if !ok {
		c.log.Errorf("RE informer returned non-RemoteEnvironment object: %#v", obj)
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(re)
	if err != nil {
		c.log.Errorf("couldn't get key for Remote Environment %#v: %v", re, err)
		return
	}
	c.log.Infof("Got RemoteEnvironment: %s", key)

	if needToAddFinalizer(re) || isDeletionCandidate(re) {
		c.queue.AddRateLimited(key)
	}
}

// Run runs the controller
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.log.Info("Starting protection controller")
	defer c.log.Info("Shutting down protection controller")

	if !controller.WaitForCacheSync("Protection", stopCh) {
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
		c.log.Errorf("Could not process RemoteEnvironment %s: %s", key, err.Error())
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	return true
}

func (c *Controller) processRE(key string) error {
	_, reName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	re, err := c.reInterface.Get(reName, v1.GetOptions{})
	if err != nil {
		return err
	}

	if needToAddFinalizer(re) {
		clone := re.DeepCopy()
		clone.ObjectMeta.Finalizers = append(clone.ObjectMeta.Finalizers, FinalizerName)
		c.log.Info("Adding finalizer")
		_, err := c.reInterface.Update(clone)
		c.log.Infof("Finalizer added")
		if err != nil {
			return err
		}
	}

	if isDeletionCandidate(re) {
		items, _ := c.emInterface.List(v1.ListOptions{})
		exists := false

		// find if environment mapping exists
		for _, item := range items.Items {
			if item.Name == reName {
				exists = true
				break
			}
		}

		// if EnvironmentMapping does not exists - remove finalizer
		if !exists {
			clone := re.DeepCopy()
			clone.ObjectMeta.Finalizers = removeString(clone.ObjectMeta.Finalizers, FinalizerName)

			c.log.Info("Removing finalizer")
			_, err := c.reInterface.Update(clone)
			c.log.Infof("Finalizer removed")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isDeletionCandidate(re *v1alpha1.RemoteEnvironment) bool {
	return re.ObjectMeta.DeletionTimestamp != nil && containsString(re.ObjectMeta.Finalizers, FinalizerName)
}

func needToAddFinalizer(re *v1alpha1.RemoteEnvironment) bool {
	return re.ObjectMeta.DeletionTimestamp == nil && !containsString(re.ObjectMeta.Finalizers, FinalizerName)
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
