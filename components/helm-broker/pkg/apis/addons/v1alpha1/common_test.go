package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecRepository_VerifyURL(t *testing.T) {
	for name, state := range map[string]struct {
		url         string
		err         bool
		expectedUrl string
		developMode bool
	}{
		"no error, develop mode": {
			url:         "http://example.com/index.yaml",
			err:         false,
			expectedUrl: "http://example.com/index.yaml",
			developMode: true,
		},
		"no error, production mode": {
			url:         "https://example.com/index.yaml",
			err:         false,
			expectedUrl: "https://example.com/index.yaml",
		},
		"no error, missing yaml file": {
			url:         "https://example.com",
			err:         false,
			expectedUrl: "https://example.com/index.yaml",
		},
		"no error, not standard yaml file": {
			url:         "https://example.com/index-testing.yaml",
			err:         false,
			expectedUrl: "https://example.com/index-testing.yaml",
		},
		"error, wrong URL": {
			url: "http//badurl",
			err: true,
		},
		"error, production mode, not secure url": {
			url: "http://example2.com/index.yaml",
			err: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Given
			sr := SpecRepository{URL: state.url}

			// When
			err := sr.VerifyURL(state.developMode)

			// Then
			if state.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, state.expectedUrl, sr.URL)
			}
		})
	}
}
