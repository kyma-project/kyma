package provider

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Client wraps the concrete getters and provide common functionality for converting the raw bundles into models.
type Client struct {
	getters             map[string]RepositoryGetter
	specifiedSchemRegex *regexp.Regexp
	log                 *logrus.Entry
	bundleLoader        addonLoader
	concreteGetter      RepositoryGetter
}

// NewClient returns new instance of Client
func NewClient(concreteGetter RepositoryGetter, bundleLoader addonLoader, log logrus.FieldLogger) (*Client, error) {
	specifiedSchemRegex, err := regexp.Compile(`^([A-Za-z0-9]+)::(.+)$`)
	if err != nil {
		return nil, err
	}
	return &Client{
		specifiedSchemRegex: specifiedSchemRegex,
		bundleLoader:        bundleLoader,
		log:                 log.WithField("service", "RepositoryGetter:CompleteBundleProvider"),
		concreteGetter:      concreteGetter,
	}, nil
}

// Cleanup calls underlying RepositoryGetter Cleanup() method
func (d *Client) Cleanup() error {
	return d.concreteGetter.Cleanup()
}

// GetCompleteAddon returns a addon with his charts as CompleteAddon instance.
func (d *Client) GetCompleteAddon(entry addon.EntryDTO) (addon.CompleteAddon, error) {
	b, c, err := d.loadBundleAndCharts(entry.Name, entry.Version)
	if err != nil {
		return addon.CompleteAddon{}, errors.Wrapf(err, "while loading bundle %v", entry.Name)
	}
	b.RepositoryURL = d.concreteGetter.BundleDocURL(entry.Name, entry.Version)

	return addon.CompleteAddon{
		Addon:  b,
		Charts: c,
	}, nil
}

// GetIndex returns all entries from given repo index
func (d *Client) GetIndex() (*addon.IndexDTO, error) {
	idxReader, err := d.concreteGetter.IndexReader()
	if err != nil {
		return nil, errors.Wrap(err, "while getting index file")
	}
	defer idxReader.Close()

	bytes, err := ioutil.ReadAll(idxReader)
	if err != nil {
		return nil, errors.Wrap(err, "while reading index file")
	}
	idx := addon.IndexDTO{}
	if err = yaml.Unmarshal(bytes, &idx); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling index file")
	}

	// Fill the proper entry name
	for name, entries := range idx.Entries {
		for idx := range entries {
			entries[idx].Name = name
		}
	}

	return &idx, nil
}

func (d *Client) loadBundleAndCharts(entryName addon.Name, version addon.Version) (*internal.Addon, []*chart.Chart, error) {
	lType, path, err := d.concreteGetter.BundleLoadInfo(entryName, version)
	if err != nil {
		return nil, nil, addon.NewFetchingError(errors.Wrapf(err, "while reading bundle archive for name [%s] and version [%v]", entryName, version))
	}

	b, charts, err := d.loadByType(lType, path)
	if err != nil {
		return nil, nil, addon.NewLoadingError(errors.Wrapf(err, "while loading bundle and charts for bundle [%s] and version [%s]", entryName, version))
	}
	return b, charts, nil
}

// LoadType define the load type of bundle located in file system
type LoadType int

const (
	// DirectoryLoadType defines that bundle should be loaded as directory
	DirectoryLoadType LoadType = iota
	// ArchiveLoadType defines that bundle should be loaded as archive (e.g. tgz)
	ArchiveLoadType LoadType = iota
	// UnknownLoadType define that bundle cannot be loaded because type is unknown
	UnknownLoadType LoadType = iota
)

func (d *Client) loadByType(loadType LoadType, path string) (*internal.Addon, []*chart.Chart, error) {
	switch loadType {
	case DirectoryLoadType:
		return d.bundleLoader.LoadDir(path)
	case ArchiveLoadType:
		reader, err := os.Open(path)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "while opening archive from path: %v", path)
		}

		b, c, err := d.bundleLoader.Load(reader)
		reader.Close()

		return b, c, err
	default:
		return nil, nil, fmt.Errorf("unsupported load type %q. Allowed load types: Directory, Archive", loadType)
	}
}
