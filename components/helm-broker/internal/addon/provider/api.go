package provider

import (
	"io"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
)

// AddonClient defines abstraction to get and unmarshal index and addon into Models
type AddonClient interface {
	GetCompleteAddon(entry addon.EntryDTO) (addon.CompleteAddon, error)
	Cleanup() error
	GetIndex() (*addon.IndexDTO, error)
}

// RepositoryGetter defines functionality for downloading addons from repository such as git, http, etc.
type RepositoryGetter interface {
	Cleanup() error
	IndexReader() (io.ReadCloser, error)
	BundleLoadInfo(name addon.Name, version addon.Version) (LoadType, string, error)
	BundleDocURL(name addon.Name, version addon.Version) string
}
