package bundle

import (
	"net/url"
	"strings"

	"time"

	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

//go:generate mockery -name=bundleSyncer -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=brokerSyncer -output=automock -outpkg=automock -case=underscore

type (
	bundleSyncer interface {
		Execute() error
		AddProvider(url string, p Provider)
		CleanProviders()
	}
	brokerSyncer interface {
		Sync(name string, maxSyncRetries int) error
	}
)

const (
	defaultMaxRetries = 5
)

// RepositoryController is responsible for updating repositories URL on configmap change
type RepositoryController struct {
	brokerName     string
	namespace      string
	brokerSyncer   brokerSyncer
	bundleSyncer   bundleSyncer
	bundleLoader   bundleLoader
	cfgMapInformer cache.SharedIndexInformer
	queue          workqueue.RateLimitingInterface
	maxRetires     int
	developMode    bool

	// testHookAsyncOpDone used only in unit tests
	testHookAsyncOpDone func()

	log *logrus.Entry
}

// NewRepositoryController returns new RepositoryController instance.
func NewRepositoryController(bundleSyncer bundleSyncer, bundleloader bundleLoader, brokerSyncer brokerSyncer, brokerName string, cfgMapInformer cache.SharedIndexInformer, log logrus.FieldLogger, devMode bool) *RepositoryController {
	c := &RepositoryController{
		bundleSyncer:   bundleSyncer,
		bundleLoader:   bundleloader,
		brokerSyncer:   brokerSyncer,
		brokerName:     brokerName,
		cfgMapInformer: cfgMapInformer,
		developMode:    devMode,

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
	var configMaps []*v1.ConfigMap
	for _, obj := range c.cfgMapInformer.GetIndexer().List() {
		cfg, ok := obj.(*v1.ConfigMap)
		if !ok {
			return fmt.Errorf("incorrect item type: %T, should be: *ConfigMap", obj)
		}
		configMaps = append(configMaps, cfg)
	}

	var existingURLs []string
	for _, cfg := range configMaps {
		existingURLs = append(existingURLs, cfg.Data["URLs"])
	}

	c.bundleSyncer.CleanProviders()
	repositories, err := c.urlsToRepositories(existingURLs)
	if err != nil {
		c.log.Warnf("Cannot create repositories for %s/%s: %s", namespace, name, err)
		return nil
	}
	for _, repoCfg := range repositories {
		repoProvider := NewProvider(NewHTTPRepository(repoCfg), c.bundleLoader, c.log.WithField("URLs", repoCfg.URL))
		c.bundleSyncer.AddProvider(repoCfg.URL, repoProvider)
	}

	if err := c.bundleSyncer.Execute(); err != nil {
		return errors.Wrap(err, "while syncing bundles")
	}
	if err := c.brokerSyncer.Sync(c.brokerName, 5); err != nil {
		return errors.Wrapf(err, "while syncing %s broker", c.brokerName)
	}

	return nil
}

func (c *RepositoryController) urlsToRepositories(urlsList []string) ([]RepositoryConfig, error) {
	var cfgs []RepositoryConfig
	var repositoryUrls []string
	urlCounter := 0

	if c.developMode {
		c.log.Info("Sysyem works on developer mode, all unsecured repository URL are allowed")
	}

	for _, urls := range urlsList {
		for _, repositoryURL := range strings.Split(urls, "\n") {
			if len(repositoryURL) < 1 {
				continue
			}
			urlCounter++
			if c.developMode {
				repositoryUrls = append(repositoryUrls, repositoryURL)
				continue
			}
			secure, err := protocolHasTLS(repositoryURL, c.developMode)
			if err != nil {
				c.log.Infof("Repository URL %q is incorrect: %s", repositoryURL, err)
				continue
			}
			if !secure {
				c.log.Infof("Repository URL %s is unsecured", repositoryURL)
				continue
			}
			repositoryUrls = append(repositoryUrls, repositoryURL)
		}
	}

	if urlCounter > 0 && len(repositoryUrls) == 0 {
		return cfgs, errors.New("All Repository URLs are incorrect or unsecured")
	}

	for _, u := range repositoryUrls {
		cfgs = append(cfgs, RepositoryConfig{
			URL: u,
		})
	}

	return cfgs, nil
}

func protocolHasTLS(repositoryURL string, developMode bool) (bool, error) {
	uri, err := url.Parse(repositoryURL)
	if err != nil {
		return false, errors.Wrap(err, "while parsing bundle repository url")
	}

	return uri.Scheme == "https", nil
}
