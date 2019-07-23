package addon

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

//go:generate mockery -name=Provider -output=automock -outpkg=automock -case=underscore

// Provider contains method which provides CompleteAddon items.
type Provider interface {
	ProvideAddons(URL string) ([]CompleteAddon, error)
}

// CompleteAddonProvider provides CompleteAddons from a repository.
type CompleteAddonProvider struct {
	log         *logrus.Entry
	addonLoader addonLoader
	repo        repository
}

// CompleteAddon aggregates a addon with his chart(s)
type CompleteAddon struct {
	Addon  *internal.Addon
	Charts []*chart.Chart
}

// ID returns the ID of the addon
func (b *CompleteAddon) ID() internal.AddonID {
	return b.Addon.ID
}

// NewProvider returns new instance of CompleteAddonProvider.
func NewProvider(repo repository, addonLoader addonLoader, log logrus.FieldLogger) *CompleteAddonProvider {
	return &CompleteAddonProvider{
		repo:        repo,
		addonLoader: addonLoader,
		log:         log.WithField("service", "addon:CompleteAddonProvider"),
	}
}

// ProvideAddons returns a list of addons with his charts as CompleteAddon instances.
// In case of addon processing errors, the won't be stopped - next addon is processed.
func (l *CompleteAddonProvider) ProvideAddons(URL string) ([]CompleteAddon, error) {
	idx, err := l.GetIndex(URL)
	if err != nil {
		return nil, err
	}

	var items []CompleteAddon
	for _, versions := range idx.Entries {
		for _, v := range versions {
			completeAddon, err := l.LoadCompleteAddon(v)
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
			items = append(items, completeAddon)
		}
	}
	l.log.Debug("Loading addons completed.")
	return items, nil
}

// LoadCompleteAddon returns a addon with his charts as CompleteAddon instances.
func (l *CompleteAddonProvider) LoadCompleteAddon(entry EntryDTO) (CompleteAddon, error) {
	addon, charts, err := l.loadAddonAndCharts(entry.Name, entry.Version)
	if err != nil {
		return CompleteAddon{}, errors.Wrapf(err, "while loading addon %v", entry.Name)
	}
	addon.RepositoryURL = l.repo.URLForAddon(entry.Name, entry.Version)

	return CompleteAddon{
		Addon:  addon,
		Charts: charts,
	}, nil
}

// GetIndex returns all entries from given repo index
func (l *CompleteAddonProvider) GetIndex(URL string) (*IndexDTO, error) {
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

func (l *CompleteAddonProvider) loadAddonAndCharts(entryName Name, version Version) (*internal.Addon, []*chart.Chart, error) {
	addonReader, err := l.repo.AddonReader(entryName, version)
	if err != nil {
		return nil, nil, NewFetchingError(errors.Wrapf(err, "while reading addon archive for name [%s] and version [%v]", entryName, version))
	}
	defer addonReader.Close()

	addon, charts, err := l.addonLoader.Load(addonReader)
	if err != nil {
		return nil, nil, NewLoadingError(errors.Wrapf(err, "while loading addon and charts for addon [%s] and version [%s]", entryName, version))
	}
	return addon, charts, nil
}
