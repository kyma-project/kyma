package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/testkit/util"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	urlFormat       = "%s/%s/v1/metadata/services"
	deleteURLFormat = "%s/%s/v1/metadata/services/%s"
)

type AppRegistryClient struct {
	httpClient *http.Client

	appRegistryURL string
	application    string
}

func NewAppRegistryClient(registryURL, application string) *AppRegistryClient {
	return &AppRegistryClient{
		appRegistryURL: registryURL,
		application:    application,
		httpClient:     &http.Client{},
	}
}

func (arc *AppRegistryClient) CreateNotSecuredAPI(t *testing.T, targetURL string) string {
	return arc.createAPI(t, arc.baseAPI(targetURL))
}

func (arc *AppRegistryClient) CreateBasicAuthSecuredAPI(t *testing.T, targetURL, username, password string) string {
	api := arc.baseAPI(targetURL).WithBasicAuth(username, password)
	return arc.createAPI(t, api)
}

func (arc *AppRegistryClient) CreateOAuthSecuredAPI(t *testing.T, targetURL, authURL, clientID, clientSecret string) string {
	api := arc.baseAPI(targetURL).WithOAuth(authURL, clientID, clientSecret)

	return arc.createAPI(t, api)
}

func (arc *AppRegistryClient) CreateNotSecuredAPICustomHeaders(t *testing.T, targetURL string, headers map[string][]string) string {
	api := arc.baseAPI(targetURL).WithCustomHeaders(&headers)

	return arc.createAPI(t, api)
}

func (arc *AppRegistryClient) CreateNotSecuredAPICustomQueryParams(t *testing.T, targetURL string, queryParams map[string][]string) string {
	api := arc.baseAPI(targetURL).WithCustomQueryParams(&queryParams)

	return arc.createAPI(t, api)
}

func (arc *AppRegistryClient) CreateAPIWithBasicAuthSecuredSpec(t *testing.T, targetURL, specURL, username, password string) string {
	api := arc.baseAPI(targetURL).WithAPISpecURL(specURL).WithBasicAuthSecuredSpec(username, password)

	return arc.createAPI(t, api)
}

func (arc *AppRegistryClient) baseAPI(targetURL string) *API {
	return &API{
		TargetUrl: targetURL,
	}
}

func (arc *AppRegistryClient) createServiceDetails(api *API) ServiceDetails {
	return ServiceDetails{
		Name:        rand.String(10),
		Description: "acceptance tests service",
		Provider:    "acc-tests",
		Api:         api,
	}
}

func (arc *AppRegistryClient) createAPI(t *testing.T, api *API) string {
	serviceDetails := arc.createServiceDetails(api)

	data, err := json.Marshal(serviceDetails)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(urlFormat, arc.appRegistryURL, arc.application), bytes.NewBuffer(data))
	require.NoError(t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(t, err)
	util.RequireStatus(t, http.StatusOK, response)

	defer response.Body.Close()

	var idResponse PostServiceResponse
	err = json.NewDecoder(response.Body).Decode(&idResponse)
	require.NoError(t, err)

	return idResponse.ID
}

func (arc *AppRegistryClient) CleanupService(t *testing.T, serviceId string) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(deleteURLFormat, arc.appRegistryURL, arc.application, serviceId), nil)
	require.NoError(t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(t, err)
	util.RequireStatus(t, http.StatusNoContent, response)
}
