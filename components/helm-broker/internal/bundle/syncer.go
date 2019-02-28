package bundle

import (
	"sync"

	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=bundleOperations -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=Provider -output=automock -outpkg=automock -case=underscore

type bundleOperations interface {
	FindAll() ([]*internal.Bundle, error)
	RemoveByID(internal.BundleID) error
	Upsert(*internal.Bundle) (replace bool, err error)
	RemoveAll() error
}

// Provider contains method which provides CompleteBundle items.
type Provider interface {
	ProvideBundles() ([]CompleteBundle, error)
}

// Syncer is responsible for loading bundles from bundle providers and stores it into the storage.
type Syncer struct {
	repositories    map[string]Provider
	bundleStorage   bundleOperations
	chartOperations storage.Chart
	log             *logrus.Entry
	mu              sync.Mutex
}

// NewSyncer returns new Syncer instance.
func NewSyncer(bundleOperations bundleOperations, chartOperations storage.Chart, log logrus.FieldLogger) *Syncer {
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

	resultChan := make(chan bundlesChange)
	for url, repo := range s.repositories {
		go func(p Provider, u string) {
			// gather bundles from external repos
			change := s.generateBundlesChange(p, u)
			resultChan <- change
		}(repo, url)
	}

	// collect results
	var toUpsert []CompleteBundle
	for range s.repositories {
		change := <-resultChan
		if !change.containsChange() {
			continue
		}

		toUpsert = append(toUpsert, change.items...)
	}

	idUsage := map[internal.BundleID][]CompleteBundle{}
	for _, item := range toUpsert {
		if _, exists := idUsage[item.ID()]; !exists {
			idUsage[item.ID()] = []CompleteBundle{}
		}
		idUsage[item.ID()] = append(idUsage[item.ID()], item)
	}

	// warn about bundles when there are more than one bundle with the same ID
	for _, bundleWithCharts := range toUpsert {
		if len(idUsage[bundleWithCharts.ID()]) > 1 {
			var msgs []string
			for _, bundleWithCharts := range idUsage[bundleWithCharts.ID()] {
				msgs = append(msgs, fmt.Sprintf("[url: %s, name: %s, version: %s]", bundleWithCharts.Bundle.Repository.URL, bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String()))
			}

			s.log.Warnf("There are more than one bundle with the same ID (%s): %s", bundleWithCharts.ID(), strings.Join(msgs, ", "))
		}
	}

	// remove bundles before upsert
	if err := s.bundleStorage.RemoveAll(); err != nil {
		return errors.Wrap(err, "while removing all bundles")
	}

	// upsert bundles
	for _, bundleWithCharts := range toUpsert {
		// do not upsert bundle if there are more than one bundle with the same ID
		if len(idUsage[bundleWithCharts.ID()]) > 1 {
			continue
		}
		if err := s.upsert(bundleWithCharts); err != nil {
			return errors.Wrapf(err, "while upsering bundle %s:%s", bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
		}
		s.log.Infof("Bundle %s:%s saved.", bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
	}

	return nil
}

func (s *Syncer) prepareIndexes(bundles []*internal.Bundle) (map[string][]*internal.Bundle, map[internal.BundleID]*internal.Bundle) {
	bundlesByRepo := map[string][]*internal.Bundle{}
	bundlesByID := map[internal.BundleID]*internal.Bundle{}
	for _, bundle := range bundles {
		url := bundle.Repository.URL
		if _, exists := bundlesByRepo[url]; !exists {
			bundlesByRepo[url] = []*internal.Bundle{}
		}
		bundlesByRepo[url] = append(bundlesByRepo[url], bundle)
		bundlesByID[bundle.ID] = bundle
	}
	return bundlesByRepo, bundlesByID
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
	// update bundles with repository url
	for _, item := range items {
		item.Bundle.Repository.URL = url
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
