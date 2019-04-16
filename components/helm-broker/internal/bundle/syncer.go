package bundle

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kyma-project/kyma/components/helm-broker/internal"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type bundleOperations interface {
	Upsert(*internal.Bundle) (replace bool, err error)
	RemoveAll() error
}

type chartUpserter interface {
	Upsert(c *chart.Chart) (replace bool, err error)
}

// Syncer is responsible for loading bundles from bundle providers and stores it into the storage.
type Syncer struct {
	repositories    map[string]Provider
	bundleStorage   bundleOperations
	chartOperations chartUpserter
	log             *logrus.Entry
	mu              sync.Mutex
}

// NewSyncer returns new Syncer instance.
func NewSyncer(bundleOperations bundleOperations, chartOperations chartUpserter, log logrus.FieldLogger) *Syncer {
	return &Syncer{
		repositories:    map[string]Provider{},
		bundleStorage:   bundleOperations,
		chartOperations: chartOperations,
		log:             log.WithField("service", "syncer"),
	}
}

// AddProvider registers new Provider which provides bundles from a repository and enrich bundles with the url.
func (s *Syncer) AddProvider(url string, provider Provider) {
	s.repositories[url] = provider
}

// CleanProviders remove all providers from the map.
func (s *Syncer) CleanProviders() {
	s.repositories = map[string]Provider{}
}

// Execute performs bundles storage with repositories synchronization.
func (s *Syncer) Execute() error {
	// Syncer must not be used in parallel
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Loading bundles from repositories")
	defer s.log.Info("Loading bundles finished")

	fetchedBundles := s.fetchBundlesFromRepositories()
	uniqueBundles := s.removeAndNotifyAboutDuplication(fetchedBundles)

	// remove bundles before upsert
	if err := s.bundleStorage.RemoveAll(); err != nil {
		return errors.Wrap(err, "while removing all bundles")
	}

	// upsert bundles
	for _, bundleWithCharts := range uniqueBundles {
		if err := s.upsert(bundleWithCharts); err != nil {
			return errors.Wrapf(err, "while upserting bundle %s:%s", bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
		}
		s.log.Infof("Bundle %s:%s saved.", bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
	}

	return nil
}

func (s *Syncer) fetchBundlesFromRepositories() []CompleteBundle {
	resultChan := make(chan bundlesChange)
	// trigger downloading
	for url, repo := range s.repositories {
		go func(p Provider, u string) {
			// gather bundles from external repos
			change := s.generateBundlesChange(p, u)
			resultChan <- change
		}(repo, url)
	}

	// collect results
	var fetchedBundles []CompleteBundle
	for range s.repositories {
		change := <-resultChan
		if !change.containsChange() {
			continue
		}

		fetchedBundles = append(fetchedBundles, change.items...)
	}

	return fetchedBundles
}

func (s *Syncer) removeAndNotifyAboutDuplication(allBundles []CompleteBundle) []CompleteBundle {
	indexedBundles := map[internal.BundleID][]CompleteBundle{}
	for _, item := range allBundles {
		if _, exists := indexedBundles[item.ID()]; !exists {
			indexedBundles[item.ID()] = []CompleteBundle{}
		}
		indexedBundles[item.ID()] = append(indexedBundles[item.ID()], item)
	}

	var uniqueBundles []CompleteBundle
	for id, bundles := range indexedBundles {
		if len(bundles) == 1 {
			uniqueBundles = append(uniqueBundles, bundles[0])
			continue
		}

		var msgs []string
		for _, duplicated := range bundles {
			msgs = append(msgs, fmt.Sprintf("[url: %s, name: %s, version: %s]", duplicated.Bundle.RemoteRepositoryURL, duplicated.Bundle.Name, duplicated.Bundle.Version.String()))
		}
		s.log.Warnf("There are more than one bundle with the same ID (%s): %s", id, strings.Join(msgs, ", "))
	}

	return uniqueBundles
}

func (s *Syncer) upsert(bundleWithCharts CompleteBundle) error {
	_, err := s.bundleStorage.Upsert(bundleWithCharts.Bundle)
	if err != nil {
		return errors.Wrapf(err, "could not upsert bundle %s (%s:%s)", bundleWithCharts.Bundle.ID, bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
	}

	for _, ch := range bundleWithCharts.Charts {
		_, err := s.chartOperations.Upsert(ch)
		if err != nil {
			return errors.Wrapf(err, "could not upsert chart %s:%s for bundle %s (%s:%s)", ch.Metadata.Name, ch.Metadata.Version, bundleWithCharts.Bundle.ID, bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
		}
	}
	return nil
}

func (s *Syncer) generateBundlesChange(repo Provider, url string) bundlesChange {
	items, err := repo.ProvideBundles()
	if err != nil {
		s.log.Errorf("Could not load bundles from the repository %s, error: %s", url, err.Error())
		return newBundlesChangeForError(err)
	}
	return bundlesChange{
		items: items,
	}
}

type bundlesChange struct {
	items []CompleteBundle
	err   error
}

// containsChange indicates the change must be processed, the simplest check returns true if there was no error.
func (bc bundlesChange) containsChange() bool {
	return bc.err == nil
}

func newBundlesChangeForError(err error) bundlesChange {
	return bundlesChange{err: err}
}
