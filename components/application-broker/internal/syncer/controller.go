package syncer

import (
	"context"
	"time"

	re_type_v1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	informers "github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

const (
	// maxRemoteEnvironmentProcessRetries is the number of times a remote environment CR will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms
	maxRemoteEnvironmentProcessRetries = 5
)

//go:generate mockery -name=remoteEnvironmentUpserter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=remoteEnvironmentRemover -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=remoteEnvironmentCRValidator -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=remoteEnvironmentCRMapper -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=scRelistRequester -output=automock -outpkg=automock -case=underscore

type (
	remoteEnvironmentUpserter interface {
		Upsert(re *internal.RemoteEnvironment) (bool, error)
	}

	remoteEnvironmentRemover interface {
		Remove(name internal.RemoteEnvironmentName) error
	}

	remoteEnvironmentCRValidator interface {
		Validate(dto *re_type_v1alpha1.RemoteEnvironment) error
	}

	remoteEnvironmentCRMapper interface {
		ToModel(dto *re_type_v1alpha1.RemoteEnvironment) *internal.RemoteEnvironment
	}

	scRelistRequester interface {
		RequestRelist()
	}
)

// Controller populates local storage with all RemoteEnvironment custom resources created in k8s cluster.
type Controller struct {
	log               logrus.FieldLogger
	queue             workqueue.RateLimitingInterface
	informer          informers.RemoteEnvironmentInformer
	reUpserter        remoteEnvironmentUpserter
	reRemover         remoteEnvironmentRemover
	reCRValidator     remoteEnvironmentCRValidator
	reCRMapper        remoteEnvironmentCRMapper
	scRelistRequester scRelistRequester
}

// New creates new remote environment controller
func New(remoteEnvironmentInformer informers.RemoteEnvironmentInformer, reUpserter remoteEnvironmentUpserter, reRemover remoteEnvironmentRemover, scRelistRequester scRelistRequester, log logrus.FieldLogger) *Controller {
	c := &Controller{
		informer:          remoteEnvironmentInformer,
		reUpserter:        reUpserter,
		reRemover:         reRemover,
		scRelistRequester: scRelistRequester,
		log:               log.WithField("service", "syncer:controller"),

		queue: workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),

		reCRValidator: &reCRValidator{},
		reCRMapper:    &reCRMapper{},
	}

	remoteEnvironmentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addRE,
		DeleteFunc: c.deleteRE,
		UpdateFunc: c.updateRE,
	})

	return c
}

func (c *Controller) addRE(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling adding event: while adding new remote environment custom resource to queue: couldn't get key: %v", err)
		return
	}

	c.queue.Add(key)
}

func (c *Controller) deleteRE(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling deletion event: while adding new remote environment custom resource to queue: couldn't get key: %v", err)
		return
	}

	c.queue.Add(key)
}

func (c *Controller) updateRE(old, cur interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err != nil {
		c.log.Errorf("while handling update event: while adding new remote environment custom resource to queue: couldn't get key: %v", err)
		return
	}

	c.queue.Add(key)
}

// Run starts the controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.shutdownQueueOnStop(stopCh)

	c.log.Info("Starting remote environment CR sync-controller")
	defer c.log.Infof("Shutting down remote environment CR sync-controller")

	if !cache.WaitForCacheSync(stopCh, c.informer.Informer().HasSynced) {
		c.log.Error("Timeout occurred on waiting for caches to sync. Shutdown the controller.")
		return
	}

	c.log.Info("RE controller synced and ready")

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

	strKey := key.(string)
	err := c.processItem(strKey)
	switch {
	case err == nil:
		c.queue.Forget(key)

		c.scRelistRequester.RequestRelist()
		c.log.Infof("Relist requested after successful processing of the %q", strKey)

	case isTemporaryError(err) && c.queue.NumRequeues(key) < maxRemoteEnvironmentProcessRetries:
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
	obj, exists, err := c.informer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return errors.Wrapf(err, "while getting object with key %q from store", key)
	}

	if !exists {
		err := c.reRemover.Remove(internal.RemoteEnvironmentName(key))
		if err != nil {
			return errors.Wrapf(err, "while removing remote environment with name %q from storage", key)
		}
		c.log.Infof("Remote Environment %q was removed from storage", key)
		return nil
	}

	reObj, ok := obj.(*re_type_v1alpha1.RemoteEnvironment)
	if !ok {
		return errors.New("cannot cast received object to v1alpha1.RemoteEnvironment type")
	}

	if err := c.reCRValidator.Validate(reObj); err != nil {
		return errors.Wrapf(err, "while validating remote environment %q", key)
	}

	dm := c.reCRMapper.ToModel(reObj)
	replaced, err := c.reUpserter.Upsert(dm)
	if err != nil {
		return errors.Wrapf(err, "while upserting remote environment with name %q into storage", key)
	}

	c.log.Infof("Remote Environment %q was added into storage (replaced: %v)", key, replaced)
	return nil
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

// isTemporaryError returns true if error implements following interface:
//	type temporary interface {
//		Temporary() bool
//	}
//
// and Temporary() method return true. Otherwise false will be returned.
func isTemporaryError(err error) bool {
	type temporary interface {
		Temporary() bool
	}

	te, ok := err.(temporary)
	return ok && te.Temporary()
}
