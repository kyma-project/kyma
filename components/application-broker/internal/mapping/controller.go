package mapping

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// maxApplicationMappingProcessRetries is the number of times a application mapping CR will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms
	maxApplicationMappingProcessRetries = 15
)

// nsBrokerFacade is responsible for managing namespaced ServiceBrokers and creating proper k8s Services for them in the system namespace
//go:generate mockery --name=nsBrokerFacade --output=automock --outpkg=automock --case=underscore --exported
type nsBrokerFacade interface {
	Create(destinationNs string) error
	Delete(destinationNs string) error
	Exist(destinationNs string) (bool, error)
}

//go:generate mockery --name=mappingLister --output=automock --outpkg=automock --case=underscore --exported
type mappingLister interface {
	ListApplicationMappings(application string) ([]*v1alpha1.ApplicationMapping, error)
}

//go:generate mockery --name=nsBrokerSyncer --output=automock --outpkg=automock --case=underscore --exported
type nsBrokerSyncer interface {
	SyncBroker(name string) error
}

type instanceChecker interface {
	AnyServiceInstanceExists(namespace string) (bool, error)
}

// Controller populates local storage with all ApplicationMapping custom resources created in k8s cluster.
type Controller struct {
	queue               workqueue.RateLimitingInterface
	emInformer          cache.SharedIndexInformer
	nsBrokerFacade      nsBrokerFacade
	nsBrokerSyncer      nsBrokerSyncer
	mappingSvc          mappingLister
	livenessCheckStatus *broker.LivenessCheckStatus
	log                 logrus.FieldLogger

	instanceChecker instanceChecker
}

// New creates new application mapping controller
func New(emInformer cache.SharedIndexInformer,
	instInformer cache.SharedIndexInformer,
	nsBrokerFacade nsBrokerFacade,
	nsBrokerSyncer nsBrokerSyncer,
	instanceChecker instanceChecker,
	log logrus.FieldLogger,
	livenessCheckStatus *broker.LivenessCheckStatus) *Controller {

	c := &Controller{
		log:                 log.WithField("service", "labeler:controller"),
		emInformer:          emInformer,
		queue:               workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		nsBrokerFacade:      nsBrokerFacade,
		nsBrokerSyncer:      nsBrokerSyncer,
		mappingSvc:          newMappingService(emInformer),
		livenessCheckStatus: livenessCheckStatus,
		instanceChecker:     instanceChecker,
	}

	// EventHandler reacts every time when we add, update or delete ApplicationMapping
	emInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addEM,
		UpdateFunc: c.updateEM,
		DeleteFunc: c.deleteEM,
	})

	instInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				c.log.Errorf("while handling deleting event: while deleting service instance resource to queue: couldn't get key: %v", err)
				return
			}
			ns, _, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				c.log.Errorf("while handling deleting event: while deleting service instance resource to queue: couldn't split key: %v", err)
				return
			}
			err = c.processRemovalInNamespace(ns)
			if err != nil {
				c.log.Errorf("while handling deleting event: while processing removal in namespace")
			}
		},
	})
	return c
}

func (c *Controller) addEM(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling adding event: while adding new application mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) updateEM(old, cur interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err != nil {
		c.log.Errorf("while handling update event: while adding new application mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) deleteEM(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling deletion event: while adding new application mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

// Run starts the controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.shutdownQueueOnStop(stopCh)

	c.log.Info("Starting Application Mappings controller")
	defer c.log.Infof("Shutting down Application Mappings controller")

	if !cache.WaitForCacheSync(stopCh, c.emInformer.HasSynced) {
		c.log.Error("Timeout occurred on waiting for informer caches to sync. Shutdown the controller.")
		return
	}
	wait.Until(c.runWorker, time.Second, stopCh)
}

// processRemovalInNamespace triggers controller to process removal an ApplicationMapping.
func (c *Controller) processRemovalInNamespace(namespace string) error {
	// put the key of non-existing object to the queue for processing.
	key, err := cache.MetaNamespaceKeyFunc(&v1alpha1.ApplicationMapping{
		ObjectMeta: v1.ObjectMeta{
			Name:      "",
			Namespace: namespace,
		},
		TypeMeta: v1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
	})
	if err != nil {
		return errors.Wrapf(err, "while adding a key to processing queue to trigger removal processing")
	}
	c.queue.Add(key)
	return nil
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

	err := c.processItem(key.(string))
	switch {
	case err == nil:
		c.queue.Forget(key)

	case c.queue.NumRequeues(key) < maxApplicationMappingProcessRetries:
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
	emObj, emExist, err := c.emInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return errors.Wrapf(err, "while getting object with key %q from the store", key)
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrapf(err, "while getting name and namespace from key %q", key)
	}

	if name == broker.LivenessApplicationSampleName {
		c.livenessCheckStatus.Succeeded = true
		c.log.Infof("livenessCheckStatus flag set to: %v", c.livenessCheckStatus.Succeeded)
		return nil
	}
	if !emExist {
		return c.ensureNsBrokerNotRegisteredIfNoMappingsOrSync(namespace)
	}

	envMapping, ok := emObj.(*v1alpha1.ApplicationMapping)
	if !ok {
		return fmt.Errorf("cannot cast received object to v1alpha1.ApplicationMapping type, type was [%T]", emObj)
	}
	return c.ensureNsBrokerRegisteredAndSynced(envMapping)
}

func (c *Controller) ensureNsBrokerRegisteredAndSynced(envMapping *v1alpha1.ApplicationMapping) error {
	brokerExist, err := c.nsBrokerFacade.Exist(envMapping.Namespace)
	if err != nil {
		return errors.Wrapf(err, "while checking if namespaced broker exist in namespace [%s]", envMapping.Namespace)
	}
	if brokerExist {
		if err = c.nsBrokerSyncer.SyncBroker(envMapping.Namespace); err != nil {
			return errors.Wrapf(err, "while syncing namespaced broker from namespace [%s]", envMapping.Namespace)
		}
		return nil
	}

	if err = c.nsBrokerFacade.Create(envMapping.Namespace); err != nil {
		return errors.Wrapf(err, "while creating namespaced broker in namespace [%s]", envMapping.Namespace)
	}

	return nil
}

func (c *Controller) ensureNsBrokerNotRegisteredIfNoMappingsOrSync(namespace string) error {
	brokerExist, err := c.nsBrokerFacade.Exist(namespace)
	if err != nil {
		return errors.Wrapf(err, "while checking if namespaced broker exist in namespace [%s]", namespace)
	}
	if !brokerExist {
		return nil
	}
	mappings, err := c.mappingSvc.ListApplicationMappings(namespace)
	if err != nil {
		return errors.Wrapf(err, "while listing application mappings from namespace [%s]", namespace)
	}
	// delete broker only if there'app no application mappings left in the namespace
	if len(mappings) > 0 {
		if err = c.nsBrokerSyncer.SyncBroker(namespace); err != nil {
			return errors.Wrapf(err, "while syncing namespaced broker from namespace [%s]", namespace)
		}
		return nil
	}

	// check if there is any application broker instance in the namespace
	existingInstance, err := c.instanceChecker.AnyServiceInstanceExists(namespace)
	if err != nil {
		return errors.Wrapf(err, "while checking instances for namespace [%s]", namespace)
	}
	if existingInstance {
		// sync broker because services are removed from the offering
		if err = c.nsBrokerSyncer.SyncBroker(namespace); err != nil {
			return errors.Wrapf(err, "while syncing namespaced broker from namespace [%s]", namespace)
		}
		return nil
	}

	if err = c.nsBrokerFacade.Delete(namespace); err != nil {
		return errors.Wrapf(err, "while removing namespaced broker from namespace [%s]", namespace)
	}
	return nil
}
