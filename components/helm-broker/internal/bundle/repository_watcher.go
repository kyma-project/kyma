package bundle

import (
	"strings"

	"github.com/knative/pkg/configmap"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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

// RepositoryWatcher is responsible for updating repositories URL on configmap change
type RepositoryWatcher struct {
	LastURL       string
	brokerName    string
	brokerSyncer  brokerSyncer
	bundleRemover bundleRemover
	bundleSyncer  bundleSyncer
	bundleLoader  bundleLoader
	reposWatcher  *configmap.InformedWatcher
	log           *logrus.Entry
}

// NewRepositoryWatcher returns new RepositoryWatcher instance.
func NewRepositoryWatcher(bundleRemover bundleRemover, bundleSyncer bundleSyncer, bundleloader bundleLoader, brokerSyncer brokerSyncer, brokerName string, reposWatcher *configmap.InformedWatcher, log logrus.FieldLogger) *RepositoryWatcher {
	return &RepositoryWatcher{
		bundleRemover: bundleRemover,
		bundleSyncer:  bundleSyncer,
		bundleLoader:  bundleloader,
		brokerSyncer:  brokerSyncer,
		brokerName:    brokerName,

		reposWatcher: reposWatcher,
		log:          log.WithField("service", "repository_watcher"),
	}
}

// StartWatchMapData starts watching at the given configmap and handle repositories update if value under a given key changed.
func (w *RepositoryWatcher) StartWatchMapData(name string, key string, stopCh chan struct{}) error {
	w.reposWatcher.Watch(name, func(configMap *v1.ConfigMap) {
		url := configMap.Data[key]

		if w.LastURL == url {
			return
		}
		if err := w.bundleRemover.RemoveAll(); err != nil {
			w.log.Errorf("Could not remove all bundles: %v", err)
			return
		}
		w.LastURL = url
		repositories := w.urlToRepositories()

		w.bundleSyncer.CleanProviders()
		for _, repoCfg := range repositories {
			repoProvider := NewProvider(NewHTTPRepository(repoCfg), w.bundleLoader, w.log.WithField(key, repoCfg.URL))
			w.bundleSyncer.AddProvider(repoCfg.URL, repoProvider)
		}
		w.bundleSyncer.Execute()

		if err := w.brokerSyncer.Sync(w.brokerName, 5); err != nil {
			w.log.Errorf("Could not synchronize the broker: %s: %v", name, err)
		}
	})

	return w.reposWatcher.Start(stopCh)
}

func (w *RepositoryWatcher) urlToRepositories() []RepositoryConfig {
	var cfgs []RepositoryConfig
	for _, url := range strings.Split(w.LastURL, ";") {
		cfgs = append(cfgs, RepositoryConfig{
			URL: url,
		})
	}
	return cfgs
}
