package provider

import (
	"io"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
)

// AddonClient defines abstraction to get and unmarshal raw index and addon into Models
type AddonClient interface {
	Cleanup() error
	GetCompleteAddon(entry addon.EntryDTO) (addon.CompleteAddon, error)
	GetIndex() (*addon.IndexDTO, error)
}

// RepositoryGetter defines functionality for downloading addons from repository such as git, http, etc.
type RepositoryGetter interface {
	Cleanup() error
	IndexReader() (io.ReadCloser, error)
	AddonLoadInfo(name addon.Name, version addon.Version) (LoadType, string, error)
	AddonDocURL(name addon.Name, version addon.Version) (string, error)
}
