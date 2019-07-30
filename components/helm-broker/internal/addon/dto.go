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
	Entries map[Name][]EntryDTO `json:"entries"`
}

// EntryDTO contains information about single addon entry
type EntryDTO struct {
	// Name is set to index entry key name
	Name Name `json:"-"`
	// DisplayName is the entry name, currently treated by us as DisplayName
	DisplayName string  `json:"name"`
	Description string  `json:"description"`
	Version     Version `json:"version"`
}

// CompleteAddon aggregates a addon with his chart(s)
type CompleteAddon struct {
	Addon  *internal.Addon
	Charts []*chart.Chart
}
