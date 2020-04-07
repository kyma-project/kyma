package syncer

import (
	"context"
	"time"

	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

const (
	// maxApplicationProcessRetries is the number of times a application CR will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms
	maxApplicationProcessRetries = 5
)

//go:generate mockery -name=applicationUpserter -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=applicationRemover -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=applicationCRValidator -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=applicationCRMapper -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=scRelistRequester -output=automock -outpkg=automock -case=underscore

type (
	applicationUpserter interface {
		Upsert(app *internal.Application) (bool, error)
	}

	applicationRemover interface {
		Remove(name internal.ApplicationName) error
	}

	applicationCRValidator interface {
		Validate(dto *appTypes.Application) error
	}

	applicationCRMapper interface {
		ToModel(dto *appTypes.Application) (*internal.Application, error)
	}

	scRelistRequester interface {
		RequestRelist()
	}
)

// Controller populates local storage with all Application custom resources created in k8s cluster.
type Controller struct {
	log               logrus.FieldLogger
	queue             workqueue.RateLimitingInterface
	informer          informers.ApplicationInformer
	appUpserter       applicationUpserter
	appRemover        applicationRemover
	appCRValidator    applicationCRValidator
	appCRMapper       applicationCRMapper
	scRelistRequester scRelistRequester
}

// New creates new application controller
func New(applicationInformer informers.ApplicationInformer, appUpserter applicationUpserter, appRemover applicationRemover, scRelistRequester scRelistRequester, log logrus.FieldLogger, apiPackagesSupport bool) *Controller {
	c := &Controller{
		informer:          applicationInformer,
		appUpserter:       appUpserter,
		appRemover:        appRemover,
		scRelistRequester: scRelistRequester,
		log:               log.WithField("service", "syncer:controller"),

		queue: workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	if apiPackagesSupport {
		c.appCRValidator = &appCRValidatorV2{}
		c.appCRMapper = &appCRMapperV2{}
	} else {
		c.appCRValidator = &appCRValidator{}
		c.appCRMapper = &appCRMapper{}
	}

	applicationInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addApp,
		DeleteFunc: c.deleteApp,
		UpdateFunc: c.updateApp,
	})

	return c
}

func (c *Controller) addApp(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling adding event: while adding new application custom resource to queue: couldn't get key: %v", err)
		return
	}

	c.queue.Add(key)
}

func (c *Controller) deleteApp(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling deletion event: while adding new application custom resource to queue: couldn't get key: %v", err)
		return
	}

	c.queue.Add(key)
}

func (c *Controller) updateApp(old, cur interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err != nil {
		c.log.Errorf("while handling update event: while adding new application custom resource to queue: couldn't get key: %v", err)
		return
	}

	c.queue.Add(key)
}

// Run starts the controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.shutdownQueueOnStop(stopCh)

	c.log.Info("Starting application CR sync-controller")
	defer c.log.Infof("Shutting down application CR sync-controller")

	if !cache.WaitForCacheSync(stopCh, c.informer.Informer().HasSynced) {
		c.log.Error("Timeout occurred on waiting for caches to sync. Shutdown the controller.")
		return
	}

	c.log.Info("Application controller synced and ready")

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

	case isTemporaryError(err) && c.queue.NumRequeues(key) < maxApplicationProcessRetries:
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
		err := c.appRemover.Remove(internal.ApplicationName(key))
		if err != nil {
			return errors.Wrapf(err, "while removing application with name %q from storage", key)
		}
		c.log.Infof("Application %q was removed from storage", key)
		return nil
	}

	app, ok := obj.(*appTypes.Application)
	if !ok {
		return errors.New("cannot cast received object to v1alpha1.Application type")
	}

	if err := c.appCRValidator.Validate(app); err != nil {
		return errors.Wrapf(err, "while validating application %q", key)
	}

	dm, err := c.appCRMapper.ToModel(app)
	if err != nil {
		errors.Wrap(err, "while mapping Application CR to model")
	}
	replaced, err := c.appUpserter.Upsert(dm)
	if err != nil {
		return errors.Wrapf(err, "while upserting application with name %q into storage", key)
	}

	c.log.Infof("Application %q was added into storage (replaced: %v)", key, replaced)
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
