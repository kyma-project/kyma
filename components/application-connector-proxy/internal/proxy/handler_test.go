package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/application-connector-proxy/internal/proxy/mocks"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	applicationName = "test-application"

	eventServicePathPrefix = "/v1/events"
	appRegistryPathPrefix  = "/v1/metadata"
)

func TestProxyHandler_ProxyAppConnectorRequests(t *testing.T) {

	certInfoHeader := `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
		`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
		`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
		`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`

	application := &v1alpha1.Application{
		ObjectMeta: v1.ObjectMeta{
			Name: applicationName,
		},
		Spec: v1alpha1.ApplicationSpec{},
	}

	t.Run("should proxy requests to Event Service and Application Registry", func(t *testing.T) {
		// given
		applicationClient := &mocks.ApplicationManager{}
		applicationClient.On("Get", applicationName, v1.GetOptions{}).Return(application, nil)

		eventServiceHandler := mux.NewRouter()
		eventServiceServer := httptest.NewServer(eventServiceHandler)
		eventServiceHost := strings.TrimPrefix(eventServiceServer.URL, "http://")

		eventServiceHandler.PathPrefix("/{application}/v1/events").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			appName := mux.Vars(r)["application"]
			assert.Equal(t, applicationName, appName)
			// TODO - more asserts
			w.WriteHeader(http.StatusOK)
		})

		appRegistryHandler := mux.NewRouter()
		appRegistryServer := httptest.NewServer(appRegistryHandler)
		appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")

		appRegistryHandler.Path("/{application}/v1/metadata/services").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			appName := mux.Vars(r)["application"]
			assert.Equal(t, applicationName, appName)
			// TODO - more asserts
			w.WriteHeader(http.StatusOK)
		})

		proxyHandler := NewProxyHandler(
			applicationClient,
			eventServicePathPrefix,
			eventServiceHost,
			appRegistryPathPrefix,
			appRegistryHost)

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/v1/events", applicationName), nil) // TODO - consider adding body
		require.NoError(t, err)
		req.Header.Set(CertificateInfoHeader, certInfoHeader)
		req = mux.SetURLVars(req, map[string]string{"application": applicationName})
		recorder := httptest.NewRecorder()

		// when
		proxyHandler.ProxyAppConnectorRequests(recorder, req)

		// then
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

}
