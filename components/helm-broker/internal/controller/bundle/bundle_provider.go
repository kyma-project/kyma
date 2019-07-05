package bundle

import (
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type BundleReader interface {
	URLForBundle(string, string) string
	IndexReader() (io.ReadCloser, error)
	BundleReader(string, string) (io.ReadCloser, error)
}

type BundleLoader interface {
	Load(in io.Reader) (*internal.Bundle, []*chart.Chart, error)
}

type BundleIndex struct {
	Entries map[string][]BundleEntry `yaml:"entries"`
}

type BundleEntry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type CompleteBundle struct {
	Bundle *internal.Bundle
	Charts []*chart.Chart
}

type BundleProvider struct {
	bundleReader BundleReader
	bundleLoader BundleLoader
	log          logrus.FieldLogger
}

func NewBundleProvider(br BundleReader, bl BundleLoader) *BundleProvider {
	return &BundleProvider{bundleReader: br, bundleLoader: bl, log: logrus.WithField("service", "bundle:loader")}
}

func (bp *BundleProvider) GetIndex() (*BundleIndex, error) {
	idxReader, err := bp.bundleReader.IndexReader()
	if err != nil {
		return nil, errors.Wrap(err, "while getting index file")
	}
	defer func() {
		err := idxReader.Close()
		if err != nil {
			bp.log.Error(err, "while closing index reader")
		}
	}()

	bytes, err := ioutil.ReadAll(idxReader)
	if err != nil {
		return nil, errors.Wrap(err, "while reading all index file")
	}
	idx := BundleIndex{}
	if err = yaml.Unmarshal(bytes, &idx); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling index file")
	}
	return &idx, nil
}

func (bp *BundleProvider) ProvideBundle(entry BundleEntry) (CompleteBundle, error) {
	bundle, charts, err := bp.loadBundleAndCharts(entry.Name, entry.Version)
	if err != nil {
		return CompleteBundle{}, NewFetchingError(errors.Wrap(err, "while loading bundle"))
	}
	bundle.RepositoryURL = bp.bundleReader.URLForBundle(entry.Name, entry.Version)

	bp.log.Info("Loading bundle completed.")
	return CompleteBundle{Bundle: bundle, Charts: charts}, nil
}

func (bp *BundleProvider) loadBundleAndCharts(entryName, version string) (*internal.Bundle, []*chart.Chart, error) {
	bundleReader, err := bp.bundleReader.BundleReader(entryName, version)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while reading bundle archive for name [%s] and version [%v]", entryName, version)
	}
	defer func() {
		err := bundleReader.Close()
		if err != nil {
			bp.log.Error(err, "while closing bundle reader")
		}
	}()

	bundle, charts, err := bp.bundleLoader.Load(bundleReader)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while loading bundle and charts for bundle [%s] and version [%s]", entryName, version)
	}

	return bundle, charts, nil
}
