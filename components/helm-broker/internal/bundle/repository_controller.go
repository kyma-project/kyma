package bundle

import (
	"strings"

	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

//go:generate mockery -name=bundleRemover -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=bundleSyncer -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=brokerSyncer -output=automock -outpkg=automock -case=underscore

type (
	bundleRemover interface {
		RemoveAll() error
	}
	bundleSyncer interface {
		Execute()
		AddProvider(url string, p Provider)
		CleanProviders()
	}
	brokerSyncer interface {
		Sync(name string, maxSyncRetries int) error
	}
)

const (
	mapLabelKey   = "repo"
	mapLabelValue = "true"

	defaultMaxRetries = 15
)

// RepositoryController is responsible for updating repositories URL on configmap change
type RepositoryController struct {
	LastUrls       []string
	brokerName     string
	namespace      string
	brokerSyncer   brokerSyncer
	bundleRemover  bundleRemover
	bundleSyncer   bundleSyncer
	bundleLoader   bundleLoader
	cfgMapInformer cache.SharedIndexInformer
	queue          workqueue.RateLimitingInterface
	maxRetires     int

	// testHookAsyncOpDone used only in unit tests
	testHookAsyncOpDone func()

	log *logrus.Entry
}

// NewRepositoryController returns new RepositoryController instance.
func NewRepositoryController(bundleRemover bundleRemover, bundleSyncer bundleSyncer, bundleloader bundleLoader, brokerSyncer brokerSyncer, brokerName string, cfgMapInformer cache.SharedIndexInformer, log logrus.FieldLogger) *RepositoryController {
	c := &RepositoryController{
		bundleRemover:  bundleRemover,
		bundleSyncer:   bundleSyncer,
		bundleLoader:   bundleloader,
		brokerSyncer:   brokerSyncer,
		brokerName:     brokerName,
		cfgMapInformer: cfgMapInformer,

		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ConfigMap"),
		maxRetires: defaultMaxRetries,

		log: log.WithField("service", "repository_watcher"),
	}
	c.cfgMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAddCfgMap,
		UpdateFunc: c.onUpdateCfgMap,
		DeleteFunc: c.onDeleteCfgMap,
	})
	return c
}

// Run starts the controller for configmaps which provides bundles repos.
func (c *RepositoryController) Run(stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		c.queue.ShutDown()
	}()

	c.log.Infof("Starting service configmap controller")
	defer c.log.Infof("Shutting down service configmap controller")

	if !cache.WaitForCacheSync(stopCh, c.cfgMapInformer.HasSynced) {
		c.log.Error("Timeout occurred on waiting for caches to sync. Shutdown the controller.")
		return
	}

	wait.Until(c.worker, time.Second, stopCh)
}

func (c *RepositoryController) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *RepositoryController) processNextWorkItem() bool {
	if c.testHookAsyncOpDone != nil {
		defer c.testHookAsyncOpDone()
	}
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key.(string))
	if err != nil {
		c.log.Errorf("Error processing %q (splitting meta namespace key failed): %v", key, err)
		c.queue.Forget(key)
		return true
	}

	retry := c.queue.NumRequeues(key)
	err = c.syncBundlesRepos(name, namespace)
	switch {
	case err == nil:
		c.queue.Forget(key)
	case retry < c.maxRetires:
		c.log.Debugf("Error processing %q (will retry - it's %d of %d): %v", key, retry, c.maxRetires, err)
		c.queue.AddRateLimited(key)

	default: // err != nil and too many retries
		c.log.Errorf("Error processing %q (giving up - to many retires): %v", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *RepositoryController) onAddCfgMap(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling addition event: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}
func (c *RepositoryController) onUpdateCfgMap(old, new interface{}) {
	oldCfgMap, ok := old.(*v1.ConfigMap)
	if !ok {
		c.log.Warnf("while handling update: cannot covert obj [%+v] of type %T to *ConfigMap", old, old)
		return
	}
	newCfgMap, ok := new.(*v1.ConfigMap)
	if !ok {
		c.log.Warnf("while handling update: cannot covert obj [%+v] of type %T to *ConfigMap", new, new)
		return
	}

	if oldCfgMap.Data["URLs"] != newCfgMap.Data["URLs"] {
		key, err := cache.MetaNamespaceKeyFunc(new)
		if err != nil {
			c.log.Errorf("while handling addition event: couldn't get key: %v", err)
			return
		}
		c.queue.Add(key)
	}
}

func (c *RepositoryController) onDeleteCfgMap(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		c.log.Errorf("while handling addition event: couldn't get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *RepositoryController) syncBundlesRepos(name string, namespace string) error {
	set, err := labels.ConvertSelectorToLabelsMap(fmt.Sprintf("%s=%s", mapLabelKey, mapLabelValue))
	if err != nil {
		return errors.Wrapf(err, "while creating label selector %s=%s", mapLabelKey, mapLabelValue)
	}

	cfgMapLister := listerv1.NewConfigMapLister(c.cfgMapInformer.GetIndexer())
	cfgs, err := cfgMapLister.List(labels.SelectorFromSet(set))
	if err != nil {
		return errors.Wrapf(err, "while listing config maps with labelSelector %q", set)
	}

	var computedURLs []string
	for _, cfg := range cfgs {
		computedURLs = append(computedURLs, cfg.Data["URLs"])
	}
	if reflect.DeepEqual(c.LastUrls, computedURLs) {
		return nil
	}
	if err := c.bundleRemover.RemoveAll(); err != nil {
		return errors.Wrapf(err, "while removing all bundles")
	}

	c.LastUrls = computedURLs
	repositories := c.lastURLToRepositories()

	c.bundleSyncer.CleanProviders()
	for _, repoCfg := range repositories {
		repoProvider := NewProvider(NewHTTPRepository(repoCfg), c.bundleLoader, c.log.WithField("URLs", repoCfg.URL))
		c.bundleSyncer.AddProvider(repoCfg.URL, repoProvider)
	}
	c.bundleSyncer.Execute()

	if err := c.brokerSyncer.Sync(c.brokerName, 5); err != nil {
		return errors.Wrapf(err, "while syncing %s broker", c.brokerName)
	}

	return nil
}

func (c *RepositoryController) lastURLToRepositories() []RepositoryConfig {
	var cfgs []RepositoryConfig
	for _, urls := range c.LastUrls {
		for _, url := range strings.Split(urls, "\n") {
			cfgs = append(cfgs, RepositoryConfig{
				URL: url,
			})
		}
	}
	return cfgs
}
