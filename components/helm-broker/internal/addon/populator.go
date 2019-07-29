package addon

import (
	"io"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type (
	// Name represents name of the Addon
	Name string
	// Version represents version of the Addon
	Version string
)

// IndexDTO contains collection of all addons from the given repository
type IndexDTO struct {
	Entries map[Name][]EntryDTO `yaml:"entries"`
}

// EntryDTO contains information about single addon entry
type EntryDTO struct {
	Name        Name    `yaml:"name"`
	Description string  `yaml:"description"`
	Version     Version `yaml:"version"`
}

type repository interface {
	IndexReader(string) (io.ReadCloser, error)
	AddonReader(name Name, version Version) (io.ReadCloser, error)
	URLForAddon(name Name, version Version) string
}

//go:generate mockery -name=addonLoader -output=automock -outpkg=automock -case=underscore
type addonLoader interface {
	Load(io.Reader) (*internal.Addon, []*chart.Chart, error)
}
