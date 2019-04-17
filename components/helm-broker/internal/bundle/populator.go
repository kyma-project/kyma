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

type indexDTO struct {
	Entries map[Name][]entryDTO `yaml:"entries"`
}

type entryDTO struct {
	Name        Name    `yaml:"name"`
	Description string  `yaml:"description"`
	Version     Version `yaml:"version"`
}

type repository interface {
	IndexReader() (io.ReadCloser, error)
	BundleReader(name Name, version Version) (io.ReadCloser, error)
	URLForBundle(name Name, version Version) string
}

//go:generate mockery -name=bundleLoader -output=automock -outpkg=automock -case=underscore
type bundleLoader interface {
	Load(io.Reader) (*internal.Bundle, []*chart.Chart, error)
}
