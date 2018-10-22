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

// GetIndexFile return name of yaml file with configuration (if not exist return default name)
func (cfg RepositoryConfig) GetIndexFile() string {
	if cfg.hasConfigFile() {
		return path.Base(cfg.BaseURL)
	}

	return "index.yaml"
}

// GetBaseURL return base url to bundles
func (cfg RepositoryConfig) GetBaseURL() string {
	if cfg.hasConfigFile() {
		return strings.TrimRight(cfg.BaseURL, cfg.GetIndexFile())
	}

	return strings.TrimRight(cfg.BaseURL, "/") + "/"
}

func (cfg RepositoryConfig) hasConfigFile() bool {
	extension := filepath.Ext(path.Base(cfg.BaseURL))

	return extension == ".yaml"
}
