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
