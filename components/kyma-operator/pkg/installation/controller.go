package installation

import (
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	kubeinformers "k8s.io/client-go/informers"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	internalClientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"
	internalscheme "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/informers/externalversions"
	listers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/listers/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/conditionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/finalizer"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	internalerrors "github.com/kyma-project/kyma/components/kyma-operator/pkg/errors"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/steps"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Installation is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a Installation
	// is synced successfully
	MessageResourceSynced = "Installation synced successfully"
)

// Controller .
type Controller struct {
	kubeClientset      *kubernetes.Clientset
	installationLister listers.InstallationLister
	installationSynced cache.InformerSynced
	queue              workqueue.RateLimitingInterface
	recorder           record.EventRecorder
	errorHandlers      internalerrors.ErrorHandlersInterface
	kymaSteps          *steps.InstallationSteps
	conditionManager   conditionmanager.Interface
	finalizerManager   *finalizer.Manager
	internalClientset  *internalClientset.Clientset
}

// NewController .
func NewController(kubeClientset *kubernetes.Clientset, kubeInformerFactory kubeinformers.SharedInformerFactory,
	internalInformerFactory informers.SharedInformerFactory, installationSteps *steps.InstallationSteps,
	conditionManager conditionmanager.Interface, finalizerManager *finalizer.Manager, internalClientset *internalClientset.Clientset) *Controller {

	installationInformer := internalInformerFactory.Installer().V1alpha1().Installations()

	internalscheme.AddToScheme(scheme.Scheme)

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "KymaOperator"})

	c := &Controller{
		kubeClientset:      kubeClientset,
		installationLister: installationInformer.Lister(),
		installationSynced: installationInformer.Informer().HasSynced,
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kymaOperatorQueue"),
		recorder:           recorder,
		errorHandlers:      &internalerrors.ErrorHandlers{},
		kymaSteps:          installationSteps,
		conditionManager:   conditionManager,
		finalizerManager:   finalizerManager,
		internalClientset:  internalClientset,
	}

	installationInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueInstallation,
		UpdateFunc: func(old, new interface{}) {
			c.enqueueInstallation(new)
		},
	})

	return c
}

// Run .
func (c *Controller) Run(stopCh <-chan struct{}) {

	defer func() {
		log.Println("Shutting down controller...")
		c.queue.ShutDown()
	}()

	// start single worker: We expect only one CR instance and only single operation is allowed at a time.
	go wait.Until(c.worker, time.Second, stopCh)

	// wait until we receive a stop signal
	<-stopCh
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
	if c.errorHandlers.CheckError("Error while parsing metadata", err) {
		return err
	}

	installation, err := c.installationLister.Installations(namespace).Get(name)
	if err != nil {

		if kubeerrors.IsNotFound(err) {
			runtime.HandleError(err)
			return nil
		}

		c.errorHandlers.LogError("Error while listing installation CRDs", err)
		return err
	}

	//Handle Delete
	if installation.IsBeingDeleted() {
		if installation.CanBeDeleted() {
			log.Println("Delete of Installation CR was requested, removing finalizer...")
			err := c.deleteFinalizer(installation)

			if c.errorHandlers.CheckError("Error while removing finalizer", err) {
				return err
			}

			return nil
		} else {
			log.Println("Delete of Installation CR was requested but it's status does not allow for it - ignoring the request")
		}
	}

	//TODO: Fill it with proper data and install UpdateStatus func
	installationData, err := config.NewInstallationData(installation)
	if c.errorHandlers.CheckError("Error while building installation data: ", err) {
		return err
	}

	if installation.ShouldInstall() {

		overrideProvider := overrides.New(c.kubeClientset)

		err = c.conditionManager.InstallStart()
		if c.errorHandlers.CheckError("Error starting install/update: ", err) {
			return err
		}

		err = c.kymaSteps.InstallKyma(installationData, overrideProvider)
		if c.errorHandlers.CheckError("Error during install/update: ", err) {
			_ = c.conditionManager.InstallError()
			return err
		}

		err = c.conditionManager.InstallSuccess()
		if c.errorHandlers.CheckError("Error finishing install/update: ", err) {
			return err
		}

	} else if installation.ShouldUninstall() {

		err = c.conditionManager.UninstallStart()
		if err != nil {
			return err
		}

		err = c.kymaSteps.UninstallKyma(installationData)
		if c.errorHandlers.CheckError("Uninstall error: ", err) {
			_ = c.conditionManager.UninstallError()
			return err
		}

		err = c.conditionManager.UninstallSuccess()
		if c.errorHandlers.CheckError("Error finishing uninstall: ", err) {
			return err
		}

	}

	c.recorder.Event(installation, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	return nil
}

func (c *Controller) deleteFinalizer(installation *v1alpha1.Installation) error {
	if !c.finalizerManager.HasFinalizer(installation) {
		return nil
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instObj, getErr := c.installationLister.Installations(installation.Namespace).Get(installation.Name)

		if getErr != nil {
			return getErr
		}

		installationCopy := instObj.DeepCopy()

		c.finalizerManager.RemoveFinalizer(installationCopy)
		_, updateErr := c.internalClientset.InstallerV1alpha1().Installations(installation.Namespace).Update(installationCopy)
		return updateErr
	})

	if retryErr != nil {
		return retryErr
	}

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

func (c *Controller) enqueueInstallation(obj interface{}) {

	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.queue.AddRateLimited(key)
}
