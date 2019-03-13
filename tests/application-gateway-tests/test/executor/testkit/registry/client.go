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
	t          *testing.T
	httpClient *http.Client

	appRegistryURL string
	application    string
}

func NewAppRegistryClient(t *testing.T, registryURL, application string) *AppRegistryClient {
	return &AppRegistryClient{
		t:              t,
		appRegistryURL: registryURL,
		application:    application,
		httpClient:     &http.Client{},
	}
}

func (arc *AppRegistryClient) CreateNotSecuredAPI(targetURL string) string {
	return arc.createAPI(arc.baseAPI(targetURL))
}

func (arc *AppRegistryClient) CreateBasicAuthSecuredAPI(targetURL, username, password string) string {
	api := arc.baseAPI(targetURL).WithBasicAuth(username, password)
	return arc.createAPI(api)
}

func (arc *AppRegistryClient) CreateOAuthSecuredAPI(targetURL, authURL, clientID, clientSecret string) string {
	api := arc.baseAPI(targetURL).WithOAuth(authURL, clientID, clientSecret)

	return arc.createAPI(api)
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

func (arc *AppRegistryClient) createAPI(api *API) string {
	serviceDetails := arc.createServiceDetails(api)

	data, err := json.Marshal(serviceDetails)
	require.NoError(arc.t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(urlFormat, arc.appRegistryURL, arc.application), bytes.NewBuffer(data))
	require.NoError(arc.t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(arc.t, err)
	util.RequireStatus(arc.t, http.StatusOK, response)

	var idResponse PostServiceResponse
	err = json.NewDecoder(response.Body).Decode(&idResponse)
	require.NoError(arc.t, err)

	return idResponse.ID
}

func (arc *AppRegistryClient) CleanupService(serviceId string) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(deleteURLFormat, arc.appRegistryURL, arc.application, serviceId), nil)
	require.NoError(arc.t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(arc.t, err)
	util.RequireStatus(arc.t, http.StatusNoContent, response)
}
