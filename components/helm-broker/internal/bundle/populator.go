package bundle

import (
	"io"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type (
	// Name represents name of the Bundle
	Name string
	// Version represents version of the Bundle
	Version string
)

// IndexDTO contains collection of all bundles from the given repository
type IndexDTO struct {
	Entries map[Name][]EntryDTO `yaml:"entries"`
}

// EntryDTO contains information about single bundle entry
type EntryDTO struct {
	Name        Name    `yaml:"name"`
	Description string  `yaml:"description"`
	Version     Version `yaml:"version"`
}

type repository interface {
	IndexReader(string) (io.ReadCloser, error)
	BundleReader(name Name, version Version) (io.ReadCloser, error)
	URLForBundle(name Name, version Version) string
}

//go:generate mockery -name=bundleLoader -output=automock -outpkg=automock -case=underscore
type bundleLoader interface {
	Load(io.Reader) (*internal.Bundle, []*chart.Chart, error)
}
