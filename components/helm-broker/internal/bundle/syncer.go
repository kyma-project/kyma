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
	FindAll() ([]*internal.Bundle, error)
	Upsert(*internal.Bundle) (replace bool, err error)
	RemoveAll() error
}

//go:generate mockery -name=docsTopicsService -output=automock -outpkg=automock -case=underscore
type docsTopicsService interface {
	EnsureClusterDocsTopic(bundle *internal.Bundle) error
	EnsureClusterDocsTopicRemoved(id string) error
}

type chartUpserter interface {
	Upsert(c *chart.Chart) (replace bool, err error)
}

// Syncer is responsible for loading bundles from bundle providers and stores it into the storage.
type Syncer struct {
	repositories      map[string]Provider
	bundleStorage     bundleOperations
	chartOperations   chartUpserter
	docsTopicsService docsTopicsService
	log               *logrus.Entry
	mu                sync.Mutex
}

// NewSyncer returns new Syncer instance.
func NewSyncer(bundleOperations bundleOperations, chartOperations chartUpserter, docsTopicsService docsTopicsService, log logrus.FieldLogger) *Syncer {
	return &Syncer{
		repositories:      map[string]Provider{},
		bundleStorage:     bundleOperations,
		chartOperations:   chartOperations,
		docsTopicsService: docsTopicsService,
		log:               log.WithField("service", "syncer"),
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

	previousBundles, err := s.bundleStorage.FindAll()
	if err != nil {
		return errors.Wrap(err, "could not load existing bundles")
	}
	newBundles := s.fetchBundlesFromRepositories()
	if err := s.deleteUnusedDocsTopics(previousBundles, newBundles); err != nil {
		return errors.Wrap(err, "while deleting unused ClusterDocsTopics")
	}
	uniqueBundles, err := s.removeAndNotifyAboutDuplication(newBundles)
	if err != nil {
		return errors.Wrap(err, "while getting unique bundles")
	}

	// remove bundles before upsert
	if err := s.bundleStorage.RemoveAll(); err != nil {
		return errors.Wrap(err, "while removing all bundles")
	}

	for _, bundleWithCharts := range uniqueBundles {
		// already we support only single spec entry
		if len(bundleWithCharts.Bundle.Docs) == 1 {
			if err := s.docsTopicsService.EnsureClusterDocsTopic(bundleWithCharts.Bundle); err != nil {
				return errors.Wrapf(err, "While ensuring ClusterDocsTopic for bundle %s: %v", bundleWithCharts.ID(), err)
			}
		}
		if err := s.upsertBundle(bundleWithCharts); err != nil {
			return errors.Wrapf(err, "while upserting bundle %s:%s", bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
		}
		s.log.Infof("Bundle %s:%s saved.", bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
	}

	return nil
}

func (s *Syncer) fetchBundlesFromRepositories() map[internal.BundleID][]CompleteBundle {
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
	fetchedBundles := make(map[internal.BundleID][]CompleteBundle)
	for range s.repositories {
		change := <-resultChan
		if !change.containsChange() {
			continue
		}

		for _, item := range change.items {
			if _, exists := fetchedBundles[item.ID()]; !exists {
				fetchedBundles[item.ID()] = []CompleteBundle{}
			}
			fetchedBundles[item.ID()] = append(fetchedBundles[item.ID()], item)
		}
	}

	return fetchedBundles
}

func (s *Syncer) removeAndNotifyAboutDuplication(newBundles map[internal.BundleID][]CompleteBundle) ([]CompleteBundle, error) {
	var uniqueBundles []CompleteBundle
	for id, bundles := range newBundles {
		if len(bundles) == 1 {
			uniqueBundles = append(uniqueBundles, bundles[0])
			continue
		}

		var msgs []string
		for _, duplicated := range bundles {
			msgs = append(msgs, fmt.Sprintf("[url: %s, name: %s, version: %s]", duplicated.Bundle.RepositoryURL, duplicated.Bundle.Name, duplicated.Bundle.Version.String()))
		}
		s.log.Warnf("There are more than one bundle with the same ID (%s): %s", id, strings.Join(msgs, ", "))

		if err := s.docsTopicsService.EnsureClusterDocsTopicRemoved(string(id)); err != nil {
			return nil, errors.Wrapf(err, "while ensuring ClusterDocsTopic %s is removed: %v", id, err)
		}
	}

	return uniqueBundles, nil
}

func (s *Syncer) deleteUnusedDocsTopics(previousBundles []*internal.Bundle, newBundles map[internal.BundleID][]CompleteBundle) error {
	for _, v := range previousBundles {
		if _, exists := newBundles[v.ID]; !exists {
			if err := s.docsTopicsService.EnsureClusterDocsTopicRemoved(string(v.ID)); err != nil {
				return errors.Wrapf(err, "while ensuring ClusterDocsTopic %s is removed", v.ID)
			}
		}
	}
	return nil
}

func (s *Syncer) upsertBundle(bundleWithCharts CompleteBundle) error {
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
