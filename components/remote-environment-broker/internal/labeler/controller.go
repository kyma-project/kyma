package labeler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// maxEnvironmentMappingProcessRetries is the number of times a environment mapping CR will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms
	maxEnvironmentMappingProcessRetries = 15
)

//go:generate mockery -name=reGetter -output=automock -outpkg=automock -case=underscore
type reGetter interface {
	Get(internal.RemoteEnvironmentName) (*internal.RemoteEnvironment, error)
}

//go:generate mockery -name=nsPatcher -output=automock -outpkg=automock -case=underscore
type nsPatcher interface {
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *corev1.Namespace, err error)
}

// Controller populates local storage with all EnvironmentMapping custom resources created in k8s cluster.
type Controller struct {
	log        logrus.FieldLogger
	queue      workqueue.RateLimitingInterface
	emInformer cache.SharedIndexInformer
	nsInformer cache.SharedIndexInformer
	nsPatcher  nsPatcher
	reGetter   reGetter
}

// New creates new environment mapping controller
func New(emInformer cache.SharedIndexInformer, nsInformer cache.SharedIndexInformer, nsPatcher nsPatcher, reGetter reGetter, log logrus.FieldLogger) *Controller {
	c := &Controller{
		log:        log.WithField("service", "labeler:controller"),
		emInformer: emInformer,
		nsInformer: nsInformer,
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		nsPatcher:  nsPatcher,
		reGetter:   reGetter,
	}

	// EventHandler reacts every time when we add, update or delete EnvironmentMapping
	emInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addEM,
		UpdateFunc: c.updateEM,
		DeleteFunc: c.deleteEM,
	})
	return c
}

func (c *Controller) addEM(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling adding event: while adding new environment mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) updateEM(old, cur interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(cur)
	if err != nil {
		c.log.Errorf("while handling update event: while adding new environment mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) deleteEM(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling deletion event: while adding new environment mapping custom resource to queue: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

// Run starts the controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	go c.shutdownQueueOnStop(stopCh)

	c.log.Info("Starting Environment Mappings controller")
	defer c.log.Infof("Shutting down Environment Mappings controller")

	if !cache.WaitForCacheSync(stopCh, c.emInformer.HasSynced) {
		c.log.Error("Timeout occurred on waiting for EM informer caches to sync. Shutdown the controller.")
		return
	}
	if !cache.WaitForCacheSync(stopCh, c.nsInformer.HasSynced) {
		c.log.Error("Timeout occurred on waiting for NS informer caches to sync. Shutdown the controller.")
		return
	}

	c.log.Info("EM controller synced and ready")

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

	err := c.processItem(key.(string))
	switch {
	case err == nil:
		c.queue.Forget(key)

	case isTemporaryError(err) && c.queue.NumRequeues(key) < maxEnvironmentMappingProcessRetries:
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
	_, exists, err := c.emInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return errors.Wrapf(err, "while getting object with key %q from the store", key)
	}

	var name, namespace string
	namespace, name, err = cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return errors.Wrapf(err, "while getting name and namespace from key %q", key)
	}

	nsObj, nsExist, nsErr := c.nsInformer.GetIndexer().GetByKey(namespace)
	if nsErr != nil || !nsExist {
		return errors.Wrapf(err, "cannot get the namespace: %q", namespace)
	}

	reNs, ok := nsObj.(*corev1.Namespace)
	if !ok {
		return errors.New("cannot cast received object to corev1.Namespace type")
	}

	if !exists {
		if err = c.deleteNsAccLabel(reNs); err != nil {
			return errors.Wrapf(err, "cannot delete AccessLabel from the namespace: %q", namespace)
		}
		return nil
	}
	var label string
	label, err = c.getReAccLabel(name)
	if err != nil {
		return errors.Wrapf(err, "cannot get AccessLabel from RE: %q", name)
	}
	err = c.applyNsAccLabel(reNs, label)
	if err != nil {
		return errors.Wrapf(err, "cannot apply AccessLabel to the namespace: %q", namespace)
	}

	return nil
}

func (c *Controller) deleteNsAccLabel(ns *corev1.Namespace) error {
	nsCopy := ns.DeepCopy()
	c.log.Infof("Deleting AccessLabel: %q, from the namespace - %q", nsCopy.Labels["accessLabel"], nsCopy.Name)

	delete(nsCopy.Labels, "accessLabel")

	err := c.patchNs(ns, nsCopy)
	if err != nil {
		return fmt.Errorf("failed to delete AccessLabel from the namespace: %q, %v", nsCopy.Name, err)
	}

	return nil
}

func (c *Controller) applyNsAccLabel(ns *corev1.Namespace, label string) error {
	nsCopy := ns.DeepCopy()
	if nsCopy.Labels == nil {
		nsCopy.Labels = make(map[string]string)
	}
	nsCopy.Labels["accessLabel"] = label

	c.log.Infof("Applying AccessLabel: %q to namespace - %q", label, nsCopy.Name)

	err := c.patchNs(ns, nsCopy)
	if err != nil {
		return fmt.Errorf("failed to apply AccessLabel: %q to the namespace: %q, %v", label, nsCopy.Name, err)
	}

	return nil
}

func (c *Controller) patchNs(nsOrig, nsMod *corev1.Namespace) error {
	oldData, err := json.Marshal(nsOrig)
	if err != nil {
		return errors.Wrapf(err, "while marshalling original namespace")
	}
	newData, err2 := json.Marshal(nsMod)
	if err2 != nil {
		return errors.Wrapf(err, "while marshalling modified namespace")
	}

	patch, err3 := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Namespace{})
	if err3 != nil {
		return errors.Wrapf(err, "while creating patch")
	}

	if _, err := c.nsPatcher.Patch(nsMod.Name, types.StrategicMergePatchType, patch); err != nil {
		return fmt.Errorf("failed to patch namespace: %q: %v", nsMod.Name, err)
	}
	return nil
}

func (c *Controller) getReAccLabel(name string) (string, error) {
	// get RE from storage
	re, err := c.reGetter.Get(internal.RemoteEnvironmentName(name))
	if err != nil {
		switch {
		// We consider IsNotFoundError as Temporary error because EM can reference to existing but not already stored RE.
		// In this case we want from Controller to retry processing this EM.
		case storage.IsNotFoundError(err):
			return "", errors.Wrapf(&tmpError{err}, "while getting remote environment with name: %q", name)
		default:
			return "", errors.Wrapf(err, "while getting remote environment with name: %q", name)
		}
	}
	if re.AccessLabel == "" {
		return "", fmt.Errorf("RE %q access label is empty", name)
	}

	return re.AccessLabel, nil
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

// and Temporary() method return true. Otherwise false will be returned.
func isTemporaryError(err error) bool {
	type temporary interface {
		Temporary() bool
	}

	te, ok := errors.Cause(err).(temporary)
	return ok && te.Temporary()
}

type tmpError struct {
	err error
}

func (t *tmpError) Error() string {
	return t.err.Error()
}

func (t *tmpError) Temporary() bool {
	return true
}
