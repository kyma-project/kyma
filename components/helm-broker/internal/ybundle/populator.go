package ybundle

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
	// BundleName represents name of the Bundle
	BundleName string
	// BundleVersion represents version of the Bundle
	BundleVersion string
)

type indexDTO struct {
	Entries map[BundleName][]entryDTO `yaml:"entries"`
}

type entryDTO struct {
	Name        BundleName    `yaml:"name"`
	Description string        `yaml:"description"`
	Version     BundleVersion `yaml:"version"`
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

	for entryName, versions := range idx.Entries {
		for _, v := range versions {
			bundleReader, bundleCloser, err := b.repo.BundleReader(entryName, v.Version)
			if err != nil {
				return errors.Wrapf(err, "while reading bundle archive for name [%s] and version [%v]", entryName, v.Version)
			}
			defer bundleCloser()

			bundle, charts, err := b.bundleLoader.Load(bundleReader)
			if err != nil {
				return errors.Wrapf(err, "while loading bundle and charts for bundle [%s] and version [%s]", entryName, v.Version)
			}

			for _, ch := range charts {
				if _, err := b.chartInterface.Upsert(ch); err != nil {
					return errors.Wrapf(err, "while storing chart [%s] for bundle [%s] with version [%s]", ch.String(), entryName, v.Version)
				}
			}

			if _, err := b.bundleInterface.Upsert(bundle); err != nil {
				return errors.Wrapf(err, "while storing bundle [%s] with version [%s]", entryName, v.Version)
			}
			b.log.Infof("Bundle with name [%s] and version [%s] successfully stored", entryName, v.Version)
		}
	}
	return nil
}

func (b *Populator) getIndex() (*indexDTO, error) {
	idxReader, idxCloser, err := b.repo.IndexReader()
	if err != nil {
		return nil, errors.Wrap(err, "while getting index.yaml")
	}
	defer idxCloser()

	bytes, err := ioutil.ReadAll(idxReader)
	if err != nil {
		return nil, errors.Wrap(err, "while reading all index.yaml")
	}
	idx := indexDTO{}
	if err = yaml.Unmarshal(bytes, &idx); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling index")
	}
	return &idx, nil
}

// be aware that after regenerating mocks, manual steps are required
//go:generate mockery -name=repository -output=automock -outpkg=automock -case=underscore
type repository interface {
	IndexReader() (r io.Reader, closer func(), err error)
	BundleReader(name BundleName, version BundleVersion) (r io.Reader, closer func(), err error)
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
