package proxy

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StripSecretFromPath(t *testing.T) {

	for _, testCase := range []struct {
		targetURL    string
		expectedPath string
	}{
		{
			targetURL:    "http://my-url.com/test",
			expectedPath: "/test",
		},
		{
			targetURL:    "http://my-url.com/test/",
			expectedPath: "/test/",
		},
		{
			targetURL:    "http://my-url.com/secret/super-secret/api/super-api",
			expectedPath: "",
		},
		{
			targetURL:    "http://my-url.com/secret/super-secret/api/super-api/",
			expectedPath: "/",
		},
		{
			targetURL:    "http://my-url.com/secret/super-secret/api/super-api/my-super-path",
			expectedPath: "/my-super-path",
		},
		{
			targetURL:    "http://my-url.com/secret/super-secret/api/super-api/my-super-path/",
			expectedPath: "/my-super-path/",
		},
		{
			targetURL:    "http://my-url.com/secret/super-secret/api/super-api/my/super/complex/path",
			expectedPath: "/my/super/complex/path",
		},
		{
			targetURL:    "http://my-url.com/secret/super-secret/api/super-api/my/super/complex/path/",
			expectedPath: "/my/super/complex/path/",
		},
	} {
		t.Run(testCase.targetURL, func(t *testing.T) {
			// when
			parsedURL, err := url.Parse(testCase.targetURL)
			require.NoError(t, err)

			path := stripSecretFromPath(parsedURL.Path)

			// then
			assert.Equal(t, testCase.expectedPath, path)
		})
	}

}
