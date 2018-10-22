package bundle_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
)

func TestRepositoryConfig_GetIndexFile_defaultValue(t *testing.T) {
	cfg := bundle.RepositoryConfig{BaseURL: "http://example.com/repository/stable/"}

	assert.Equal(t, "index.yaml", cfg.GetIndexFile())
}

func TestRepositoryConfig_GetIndexFile(t *testing.T) {
	cfg := bundle.RepositoryConfig{BaseURL: "http://example.com/repository/stable/conf.yaml"}

	assert.Equal(t, "conf.yaml", cfg.GetIndexFile())
}

func TestRepositoryConfig_GetBaseUrl(t *testing.T) {
	cfg := bundle.RepositoryConfig{BaseURL: "http://example.com/repository/stable/index.yaml"}

	assert.Equal(t, "http://example.com/repository/stable/", cfg.GetBaseURL())
}

func TestRepositoryConfig_GetBaseUrl_withoutYamlFile(t *testing.T) {
	cfg := bundle.RepositoryConfig{BaseURL: "http://example.com/repository/stable/"}

	assert.Equal(t, "http://example.com/repository/stable/", cfg.GetBaseURL())
}

func TestRepositoryConfig_GetBaseUrl_withIncompletePath(t *testing.T) {
	cfg := bundle.RepositoryConfig{BaseURL: "http://example.com/repository/stable"}

	assert.Equal(t, "http://example.com/repository/stable/", cfg.GetBaseURL())
}
