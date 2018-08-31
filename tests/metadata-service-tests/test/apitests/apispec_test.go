package apitests

import (
	"github.com/kyma-project/kyma/tests/metadata-service-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"testing"
)

func TestApiSpec(t *testing.T) {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sInClusterResourcesClient(config.Namespace)
	require.NoError(t, err)

	dummyRE, err := k8sResourcesClient.CreateDummyRemoteEnvironment("dummy-re", v1.GetOptions{})
	require.NoError(t, err)

	t.Run("Application Connector Metadata", func(t *testing.T) {

		t.Run("should return api spec", func(t *testing.T) {
			// given
			url := config.MetadataServiceUrl +"/" + dummyRE.Name + "/v1/metadataapi.yaml"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// when
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, response.StatusCode)

			rawApiSpec, err := ioutil.ReadAll(response.Body)
			require.NoError(t, err)

			var apiSpec struct{}
			err = yaml.Unmarshal(rawApiSpec, &apiSpec)
			require.NoError(t, err)
		})

		t.Run("should redirect to api spec url", func(t *testing.T) {
			// given
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					require.Equal(t, "/"+dummyRE.Name+"/v1/metadataapi.yaml", req.URL.Path)
					return http.ErrUseLastResponse
				},
			}

			url := config.MetadataServiceUrl + "/" + dummyRE.Name + "/v1/metadata"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// when
			response, err := client.Do(request)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusMovedPermanently, response.StatusCode)
		})
	})

	err = k8sResourcesClient.DeleteRemoteEnvironment(dummyRE.Name, &v1.DeleteOptions{})
	require.NoError(t, err)
}
