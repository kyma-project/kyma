package apitests

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/application-registry-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiSpec(t *testing.T) {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sInClusterResourcesClient(config.Namespace)
	require.NoError(t, err)

	dummyApp, err := k8sResourcesClient.CreateDummyApplication("appapispectest0", v1.GetOptions{}, true)
	require.NoError(t, err)

	t.Run("Application Connector Metadata", func(t *testing.T) {

		t.Run("should return api spec", func(t *testing.T) {
			// given
			url := config.MetadataServiceUrl + "/" + dummyApp.Name + "/v1/metadata/api.yaml"

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
					require.Equal(t, "/"+dummyApp.Name+"/v1/metadata/api.yaml", req.URL.Path)
					return http.ErrUseLastResponse
				},
			}

			url := config.MetadataServiceUrl + "/" + dummyApp.Name + "/v1/metadata"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// when
			response, err := client.Do(request)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusMovedPermanently, response.StatusCode)
		})
	})

	err = k8sResourcesClient.DeleteApplication(dummyApp.Name, &v1.DeleteOptions{})
	require.NoError(t, err)
}
