package release

import (
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	internalscheme "github.com/kyma-project/kyma/components/installer/pkg/client/clientset/versioned/scheme"
	informers "github.com/kyma-project/kyma/components/installer/pkg/client/informers/externalversions"
	listers "github.com/kyma-project/kyma/components/installer/pkg/client/listers/release/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/finalizer"

	internalerrors "github.com/kyma-project/kyma/components/installer/pkg/errors"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Release is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a Release
	// is synced successfully
	MessageResourceSynced = "Release synced successfully"
)

// Controller is the controller implementation for Release resources
type Controller struct {
	kubeClientset    *kubernetes.Clientset
	releaseLister    listers.ReleaseLister
	queue            workqueue.RateLimitingInterface
	recorder         record.EventRecorder
	errorHandlers    internalerrors.ErrorHandlersInterface
	finalizerManager *finalizer.Manager
}

// NewController returns a new instance of release Controller
func NewController(kubeClientset *kubernetes.Clientset, internalInformerFactory informers.SharedInformerFactory, finalizerManager *finalizer.Manager) *Controller {

	releaseInformer := internalInformerFactory.Release().V1alpha1().Releases()

	internalscheme.AddToScheme(scheme.Scheme)

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "kymaInstaller"})

	c := &Controller{
		kubeClientset:    kubeClientset,
		releaseLister:    releaseInformer.Lister(),
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kymaReleaseQueue"),
		recorder:         recorder,
		errorHandlers:    &internalerrors.ErrorHandlers{},
		finalizerManager: finalizerManager,
	}

	releaseInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		//TBD
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(old, new interface{}) {},
		DeleteFunc: func(obv interface{}) {},
	})

	return c
}

// Run runs the controller
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {

	defer func() {
		log.Println("Shutting down Release controller...")
		c.queue.ShutDown()
	}()

	for i := 0; i < workers; i++ {
		//start workers
		go wait.Until(c.worker, time.Second, stopCh)
	}
}

func (c *Controller) worker() {

	// process until we're told to stop
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {

	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.syncHandler(key.(string))
	c.handleErr(err, key)
	return true
}

func (c *Controller) syncHandler(key string) error {

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	release, err := c.releaseLister.Releases(namespace).Get(name)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			runtime.HandleError(err)
			return nil
		}

		return err
	}

	// Check if deletion has been triggered
	// with ReleaseFinalizerHandler

	c.recorder.Event(release, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) handleErr(err error, key interface{}) {

	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	runtime.HandleError(err)
}
