package bundle

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/sirupsen/logrus"
)

//go:generate mockery -name=bundleOperations -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=provider -output=automock -outpkg=automock -case=underscore

type bundleOperations interface {
	FindAll() ([]*internal.Bundle, error)
	RemoveByID(internal.BundleID) error
	Upsert(*internal.Bundle) (replace bool, err error)
}

// provider contains method which provides CompleteBundle items.
type provider interface {
	ProvideBundles() ([]CompleteBundle, error)
}

// Syncer is responsible for loading bundles from bundle providers and stores it into the storage.
type Syncer struct {
	repositories    map[string]provider
	bundleStorage   bundleOperations
	chartOperations storage.Chart
	log             *logrus.Entry
	mu              sync.Mutex
}

// NewSyncer returns new Syncer instance.
func NewSyncer(bundleOperations bundleOperations, chartOperations storage.Chart, log logrus.FieldLogger) *Syncer {
	return &Syncer{
		repositories:    map[string]provider{},
		bundleStorage:   bundleOperations,
		chartOperations: chartOperations,
		log:             log.WithField("service", "syncer"),
	}
}

// AddProvider registers new provider which provides bundles from a repository and enrich bundles with the url.
func (s *Syncer) AddProvider(url string, provider provider) {
	s.repositories[url] = provider
}

// Execute performs bundles storage with repositories synchronization.
func (s *Syncer) Execute() {
	// Syncer must not be used in parallel
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Infof("Loading bundles from repositories")
	defer s.log.Infof("Loading bundles finished")

	bundles, err := s.bundleStorage.FindAll()
	if err != nil {
		s.log.Errorf("Could not load existing bundles, error: %s", err)
		return
	}

	bundlesByRepo, bundlesByID := s.prepareIndexes(bundles)

	resultChan := make(chan bundlesChange)
	for url, repo := range s.repositories {
		go func(p provider, b []*internal.Bundle, u string) {
			// gather bundles from external repos
			change := s.generateBundlesChange(p, b, u)
			resultChan <- change
		}(repo, bundlesByRepo[url], url)
	}

	// collect results
	var idsToRemove []internal.BundleID
	var toUpsert []CompleteBundle
	for range s.repositories {
		change := <-resultChan
		if !change.containsChange() {
			continue
		}

		idsToRemove = append(idsToRemove, change.idsToRemove...)
		toUpsert = append(toUpsert, change.items...)
	}

	idUsage := map[internal.BundleID][]CompleteBundle{}
	for _, item := range toUpsert {
		if _, exists := idUsage[item.ID()]; !exists {
			idUsage[item.ID()] = []CompleteBundle{}
		}
		idUsage[item.ID()] = append(idUsage[item.ID()], item)
	}

	// Remove all bundles which must be removed, then upsert new ones

	// Remove bundles which were deleted
	for _, id := range idsToRemove {
		s.removeBundle(id, bundlesByID[id])
	}

	// Remove bundles when there are more than one bundle with the same ID
	for _, bundleWithCharts := range toUpsert {
		if len(idUsage[bundleWithCharts.ID()]) > 1 {
			// when loaded bundles contains more than one CompleteBundle with the same ID
			// - such CompleteBundle must not exist after the sync.
			// If such CompleteBundle was existing before - remove it.
			for _, existingBundle := range bundles {
				if existingBundle.ID == bundleWithCharts.ID() {
					s.removeBundle(existingBundle.ID, existingBundle)
				}
			}

			var msgs []string
			for _, bundleWithCharts := range idUsage[bundleWithCharts.ID()] {
				msgs = append(msgs, fmt.Sprintf("[url: %s, name: %s, version: %s]", bundleWithCharts.Bundle.Repository.URL, bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String()))
			}

			s.log.Warnf("There are more than one bundle with the same ID (%s): %s", bundleWithCharts.ID(), strings.Join(msgs, ", "))
		}
	}

	// Upsert bundles
	for _, bundleWithCharts := range toUpsert {
		// do not upsert bundle if there are more than one bundle with the same ID
		if len(idUsage[bundleWithCharts.ID()]) > 1 {
			continue
		}
		s.upsert(bundleWithCharts)
		s.log.Infof("Bundle %s:%s saved.", bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version.String())
	}
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

func (s *Syncer) upsert(bundleWithCharts CompleteBundle) {
	_, err := s.bundleStorage.Upsert(bundleWithCharts.Bundle)
	if err != nil {
		s.log.Errorf("Could not upsert bundle %s (%s:%s): %s", bundleWithCharts.Bundle.ID, bundleWithCharts.Bundle.Name, bundleWithCharts.Bundle.Version, err.Error())
	}

	for _, ch := range bundleWithCharts.Charts {
		_, err := s.chartOperations.Upsert(ch)
		if err != nil {
			s.log.Errorf("Could not upsert chart %s:%s for bundle %s (%s:%s): %s")
		}
	}
}

func (s *Syncer) removeBundle(id internal.BundleID, existingBundle *internal.Bundle) {
	err := s.bundleStorage.RemoveByID(id)
	if err != nil {
		s.log.Warnf("Could not remove bundle %s from the storage: %s", id, err.Error())
		// if the removal failed, do not remove its charts
		return
	}

	if existingBundle != nil {
		for _, plan := range existingBundle.Plans {
			err := s.chartOperations.Remove(plan.ChartRef.Name, plan.ChartRef.Version)
			if err != nil {
				s.log.Warnf("Could not remove chart %s:%s used by bundle id: %s name: %s", plan.ChartRef.Name, plan.ChartRef.Version, id, existingBundle.Name)
			}
		}
	}
}

func (s *Syncer) generateBundlesChange(repo provider, existingBundles []*internal.Bundle, url string) bundlesChange {
	items, err := repo.ProvideBundles()
	if err != nil {
		s.log.Errorf("Could not load bundles from the repository %s, error: %s", url, err.Error())
		return newBundlesChangeForError(err)
	}

	// calculate removed
	var idsToRemove []internal.BundleID

	// find all bundle IDs to remove
	for _, b := range existingBundles {
		bundleRemoved := true

		// if the bundle exists in the newBundles collection - bundle must not be removed
		for _, bundleWithCharts := range items {
			if b.ID == bundleWithCharts.ID() {
				bundleRemoved = false
				break
			}
		}
		if bundleRemoved {
			idsToRemove = append(idsToRemove, b.ID)
		}
	}

	// update bundles with repository url
	for _, item := range items {
		item.Bundle.Repository.URL = url
	}

	return bundlesChange{
		idsToRemove: idsToRemove,
		items:       items,
	}
}

type bundlesChange struct {
	idsToRemove []internal.BundleID
	items       []CompleteBundle
	err         error
}

// containsChange indicates the change must be processed, the simplest check returns true if there was no error.
func (bc bundlesChange) containsChange() bool {
	return bc.err == nil
}

func newBundlesChangeForError(err error) bundlesChange {
	return bundlesChange{err: err}
}
