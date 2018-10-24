package bundle_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
)

func TestRepositoryConfig(t *testing.T) {
	tests := map[string]struct {
		givenRepoURL string

		expIndexName string
		expBaseURL   string
	}{
		"index not provided in URL": {
			givenRepoURL: "http://example.com/repository/stable/",

			expIndexName: "index.yaml",
			expBaseURL:   "http://example.com/repository/stable/",
		},
		"custom index provided in URL": {
			givenRepoURL: "http://example.com/repository/stable/conf.yaml",

			expIndexName: "conf.yaml",
			expBaseURL:   "http://example.com/repository/stable/",
		},
		"index provided in URL same as default one": {
			givenRepoURL: "http://example.com/repository/stable/index.yaml",

			expIndexName: "index.yaml",
			expBaseURL:   "http://example.com/repository/stable/",
		},
		"trailing slash is always added for BaseURL": {
			givenRepoURL: "http://example.com/repository/stable",

			expIndexName: "index.yaml",
			expBaseURL:   "http://example.com/repository/stable/",
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			cfg := bundle.RepositoryConfig{URL: tc.givenRepoURL}

			assert.Equal(t, tc.expBaseURL, cfg.BaseURL())
			assert.Equal(t, tc.expIndexName, cfg.IndexFileName())

		})
	}
}
