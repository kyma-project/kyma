package bundle

import (
	"path"
	"path/filepath"
	"strings"
)

// RepositoryConfig provides configuration for HTTP Repository.
type RepositoryConfig struct {
	BaseURL string `json:"baseUrl" valid:"required"`
}

// IndexFileName returns name of yaml file with configuration (if not exist return default name)
func (cfg RepositoryConfig) IndexFileName() string {
	if cfg.hasConfigFile() {
		return path.Base(cfg.BaseURL)
	}

	return "index.yaml"
}

// GetBaseURL returns base url to bundles with trailing slash
func (cfg RepositoryConfig) GetBaseURL() string {
	if cfg.hasConfigFile() {
		return strings.TrimRight(cfg.BaseURL, cfg.IndexFileName())
	}

	return strings.TrimRight(cfg.BaseURL, "/") + "/"
}

func (cfg RepositoryConfig) hasConfigFile() bool {
	extension := filepath.Ext(path.Base(cfg.BaseURL))

	return extension == ".yaml"
}
