package bundle

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

//go:generate mockery -name=Provider -output=automock -outpkg=automock -case=underscore

// Provider contains method which provides CompleteBundle items.
type Provider interface {
	ProvideBundles(URL string) ([]CompleteBundle, error)
}

// CompleteBundleProvider provides CompleteBundles from a repository.
type CompleteBundleProvider struct {
	log          *logrus.Entry
	bundleLoader bundleLoader
	repo         repository
}

// CompleteBundle aggregates a bundle with his chart(s)
type CompleteBundle struct {
	Bundle *internal.Bundle
	Charts []*chart.Chart
}

// ID returns the ID of the bundle
func (b *CompleteBundle) ID() internal.BundleID {
	return b.Bundle.ID
}

// NewProvider returns new instance of CompleteBundleProvider.
func NewProvider(repo repository, bundleLoader bundleLoader, log logrus.FieldLogger) *CompleteBundleProvider {
	return &CompleteBundleProvider{
		repo:         repo,
		bundleLoader: bundleLoader,
		log:          log.WithField("service", "bundle:CompleteBundleProvider"),
	}
}

// ProvideBundles returns a list of bundles with his charts as CompleteBundle instances.
// In case of bundle processing errors, the won't be stopped - next bundle is processed.
func (l *CompleteBundleProvider) ProvideBundles(URL string) ([]CompleteBundle, error) {
	idx, err := l.GetIndex(URL)
	if err != nil {
		return nil, err
	}

	var items []CompleteBundle
	for _, versions := range idx.Entries {
		for _, v := range versions {
			completeBundle, err := l.LoadCompleteBundle(v)
			switch {
			case err == nil:
			case IsFetchingError(err):
				l.log.Warnf("detected fetching problem: %v", err)
				continue
			case IsLoadingError(err):
				l.log.Warnf("detected loading problem: %v", err)
				continue
			default:
				l.log.Warnf("detected internal problem: %v", err)
				continue
			}
			items = append(items, completeBundle)
		}
	}
	l.log.Debug("Loading bundles completed.")
	return items, nil
}

// LoadCompleteBundle returns a bundle with his charts as CompleteBundle instances.
func (l *CompleteBundleProvider) LoadCompleteBundle(entry EntryDTO) (CompleteBundle, error) {
	bundle, charts, err := l.loadBundleAndCharts(entry.Name, entry.Version)
	if err != nil {
		return CompleteBundle{}, errors.Wrapf(err, "while loading bundle %v", entry.Name)
	}
	bundle.RepositoryURL = l.repo.URLForBundle(entry.Name, entry.Version)

	return CompleteBundle{
		Bundle: bundle,
		Charts: charts,
	}, nil
}

// GetIndex returns all entries from given repo index
func (l *CompleteBundleProvider) GetIndex(URL string) (*IndexDTO, error) {
	idxReader, err := l.repo.IndexReader(URL)
	if err != nil {
		return nil, errors.Wrap(err, "while getting index file")
	}
	defer idxReader.Close()

	bytes, err := ioutil.ReadAll(idxReader)
	if err != nil {
		return nil, errors.Wrap(err, "while reading all index file")
	}
	idx := IndexDTO{}
	if err = yaml.Unmarshal(bytes, &idx); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling index file")
	}
	return &idx, nil
}

func (l *CompleteBundleProvider) loadBundleAndCharts(entryName Name, version Version) (*internal.Bundle, []*chart.Chart, error) {
	bundleReader, err := l.repo.BundleReader(entryName, version)
	if err != nil {
		return nil, nil, NewFetchingError(errors.Wrapf(err, "while reading bundle archive for name [%s] and version [%v]", entryName, version))
	}
	defer bundleReader.Close()

	bundle, charts, err := l.bundleLoader.Load(bundleReader)
	if err != nil {
		return nil, nil, NewLoadingError(errors.Wrapf(err, "while loading bundle and charts for bundle [%s] and version [%s]", entryName, version))
	}
	return bundle, charts, nil
}
