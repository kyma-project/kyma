package apitests

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/connector-service-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestApiSpec(t *testing.T) {

	apiSpecPath := "/v1/api.yaml"

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	hc := testkit.NewHttpClient(config.SkipSslVerify)

	if !config.Compass {
		t.Run("should receive api spec", func(t *testing.T) {
			// when
			response, err := hc.Get(config.ExternalAPIUrl + apiSpecPath)

			// then
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)

			// when
			body, err := ioutil.ReadAll(response.Body)

			// then
			require.NoError(t, err)

			var apiSpec struct{}
			err = yaml.Unmarshal(body, &apiSpec)
			require.NoError(t, err)
		})

		t.Run("should receive 301 when accessing base path", func(t *testing.T) {
			// given
			hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				require.Equal(t, apiSpecPath, req.URL.Path)
				return http.ErrUseLastResponse
			}

			// when
			response, err := hc.Get(config.ExternalAPIUrl + "/v1")

			// then
			require.NoError(t, err)
			require.Equal(t, http.StatusMovedPermanently, response.StatusCode)
		})
	}
}
