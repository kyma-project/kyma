package bundle

import (
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type (
	// Name represents name of the Bundle
	Name string
	// Version represents version of the Bundle
	Version string
)

type indexDTO struct {
	Entries map[Name][]entryDTO `yaml:"entries"`
}

type entryDTO struct {
	Name        Name    `yaml:"name"`
	Description string  `yaml:"description"`
	Version     Version `yaml:"version"`
}

// Populator is responsible for populating bundles and charts into storage.
// Source data is provided by bundleLoader.
type Populator struct {
	repo            repository
	bundleLoader    bundleLoader
	bundleInterface bundleUpserter
	chartInterface  chartUpserter
	log             *logrus.Entry
}

// NewPopulator creates new instance of Populator.
func NewPopulator(p repository, bundleLoader bundleLoader, bundleInterface bundleUpserter, chartInterface chartUpserter, log *logrus.Entry) *Populator {
	return &Populator{
		repo:            p,
		bundleLoader:    bundleLoader,
		bundleInterface: bundleInterface,
		chartInterface:  chartInterface,
		log:             log.WithField("service", "populator"),
	}
}

// Init triggers population process.
func (b *Populator) Init() error {
	idx, err := b.getIndex()
	if err != nil {
		return err
	}

	loaded := 0
	failed := 0
	for entryName, versions := range idx.Entries {
		for _, v := range versions {
			err := b.loadBundle(entryName, v.Version)
			if err != nil {
				failed = failed + 1
				b.log.Warnf("Could not load bundle: %s", err.Error())
			} else {
				loaded = loaded + 1
			}
		}
	}
	b.log.Infof("Loading bundles completed. Successfully loaded %d bundles, failed %d bundles.", loaded, failed)
	return nil
}

func (b *Populator) loadBundle(entryName Name, version Version) error {
	bundleReader, bundleCloser, err := b.repo.BundleReader(entryName, version)
	if err != nil {
		return errors.Wrapf(err, "while reading bundle archive for name [%s] and version [%v]", entryName, version)
	}
	defer bundleCloser()

	bundle, charts, err := b.bundleLoader.Load(bundleReader)
	if err != nil {
		return errors.Wrapf(err, "while loading bundle and charts for bundle [%s] and version [%s]", entryName, version)
	}

	for _, ch := range charts {
		if _, err := b.chartInterface.Upsert(ch); err != nil {
			return errors.Wrapf(err, "while storing chart [%s] for bundle [%s] with version [%s]", ch.String(), entryName, version)
		}
	}

	if _, err := b.bundleInterface.Upsert(bundle); err != nil {
		return errors.Wrapf(err, "while storing bundle [%s] with version [%s]", entryName, version)
	}
	b.log.Infof("Bundle with name [%s] and version [%s] successfully stored", entryName, version)

	return nil
}

func (b *Populator) getIndex() (*indexDTO, error) {
	idxReader, idxCloser, err := b.repo.IndexReader()
	if err != nil {
		return nil, errors.Wrap(err, "while getting index file")
	}
	defer idxCloser()

	bytes, err := ioutil.ReadAll(idxReader)
	if err != nil {
		return nil, errors.Wrap(err, "while reading all index file")
	}
	idx := indexDTO{}
	if err = yaml.Unmarshal(bytes, &idx); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling index file")
	}
	return &idx, nil
}

// be aware that after regenerating mocks, manual steps are required
//go:generate mockery -name=repository -output=automock -outpkg=automock -case=underscore
type repository interface {
	IndexReader() (r io.Reader, closer func(), err error)
	BundleReader(name Name, version Version) (r io.Reader, closer func(), err error)
}

//go:generate mockery -name=bundleUpserter -output=automock -outpkg=automock -case=underscore
type bundleUpserter interface {
	Upsert(*internal.Bundle) (replace bool, err error)
}

//go:generate mockery -name=chartUpserter -output=automock -outpkg=automock -case=underscore
type chartUpserter interface {
	Upsert(c *chart.Chart) (replace bool, err error)
}

//go:generate mockery -name=bundleLoader -output=automock -outpkg=automock -case=underscore
type bundleLoader interface {
	Load(io.Reader) (*internal.Bundle, []*chart.Chart, error)
}
