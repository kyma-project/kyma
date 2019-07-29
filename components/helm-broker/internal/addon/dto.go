package addon

import (
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
	// Name is set to index entry key name
	Name Name
	// DisplayName is the entry name, currently treated by us as DisplayName
	DisplayName string  `yaml:"name"`
	Description string  `yaml:"description"`
	Version     Version `yaml:"version"`
}

// CompleteAddon aggregates a bundle with his chart(s)
type CompleteAddon struct {
	Addon  *internal.Addon
	Charts []*chart.Chart
}
