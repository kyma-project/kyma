package validationproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorilla/mux"
)

const (
	applicationName = "test-application"

	group  = "group"
	tenant = "tenant"

	eventServicePathPrefixV1 = "/test-application/v1/events"
	eventServicePathPrefixV2 = "/test-application/v2/events"
	appRegistryPathPrefix    = "/test-application/v1/metadata"
)

type event struct {
	Title string `json:"title"`
}

func TestProxyHandler_ProxyAppConnectorRequests(t *testing.T) {

	eventServiceHandler := mux.NewRouter()
	eventServiceServer := httptest.NewServer(eventServiceHandler)
	eventServiceHost := strings.TrimPrefix(eventServiceServer.URL, "http://")

	appRegistryHandler := mux.NewRouter()
	appRegistryServer := httptest.NewServer(appRegistryHandler)
	appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")

	testCases := []struct {
		caseDescription string
		tenant          string
		group           string
		certInfoHeader  string
		expectedStatus  int
	}{
		{
			caseDescription: "Application without group and tenant",
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusOK,
		},
		{
			caseDescription: "Application without group and tenant and with invalid Common Name",
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=invalid-cn,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
		},
		{
			caseDescription: "Application with group and tenant",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=group,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusOK,
		},
		{
			caseDescription: "Application with group, tenant and invalid Common Name",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=invalid-application,OU=group,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
		},
		{
			caseDescription: "Application with group, tenant and invalid Organizational Unit",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=invalid,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
		},
		{
			caseDescription: "Application with group, tenant and invalid Organization",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=group,O=invalid,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
		},
		{
			caseDescription: "X-Forwarded-Client-Cert header not specified",
			tenant:          tenant,
			group:           group,
			expectedStatus:  http.StatusInternalServerError,
		},
	}

	t.Run("should proxy requests", func(t *testing.T) {
		for _, testCase := range testCases {
			// given
			proxyHandler := NewProxyHandler(
				testCase.group,
				testCase.tenant,
				eventServicePathPrefixV1,
				eventServicePathPrefixV2,
				eventServiceHost,
				appRegistryPathPrefix,
				appRegistryHost)

			t.Run("should proxy event service V1 request when "+testCase.caseDescription, func(t *testing.T) {
				eventTitle := "my-event-1"

				eventServiceHandler.PathPrefix("/{application}/v1/events").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					appName := mux.Vars(r)["application"]
					assert.Equal(t, applicationName, appName)

					var receivedEvent event

					err := json.NewDecoder(r.Body).Decode(&receivedEvent)
					require.NoError(t, err)
					assert.Equal(t, eventTitle, receivedEvent.Title)

					w.WriteHeader(http.StatusOK)
				})

				body, err := json.Marshal(event{Title: eventTitle})
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/v1/events", applicationName), bytes.NewReader(body))
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": applicationName})
				recorder := httptest.NewRecorder()

				// when
				proxyHandler.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})

			t.Run("should proxy event service V2 request when "+testCase.caseDescription, func(t *testing.T) {
				eventTitle := "my-event-2"

				eventServiceHandler.PathPrefix("/{application}/v2/events").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					appName := mux.Vars(r)["application"]
					assert.Equal(t, applicationName, appName)

					var receivedEvent event

					err := json.NewDecoder(r.Body).Decode(&receivedEvent)
					require.NoError(t, err)
					assert.Equal(t, eventTitle, receivedEvent.Title)

					w.WriteHeader(http.StatusOK)
				})

				body, err := json.Marshal(event{Title: eventTitle})
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/v2/events", applicationName), bytes.NewReader(body))
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": applicationName})
				recorder := httptest.NewRecorder()

				// when
				proxyHandler.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})

			t.Run("should proxy application registry request when "+testCase.caseDescription, func(t *testing.T) {
				appRegistryHandler.PathPrefix("/{application}/v1/metadata/services").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					appName := mux.Vars(r)["application"]
					assert.Equal(t, applicationName, appName)
					w.WriteHeader(http.StatusOK)
				})

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/v1/metadata/services", applicationName), nil)
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": applicationName})
				recorder := httptest.NewRecorder()

				// when
				proxyHandler.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})
		}
	})

	t.Run("should return 400 when application not specified in path", func(t *testing.T) {
		// given
		certInfoHeader :=
			`Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";URI=`

		proxyHandler := NewProxyHandler(
			group,
			tenant,
			eventServicePathPrefixV1,
			eventServicePathPrefixV2,
			eventServiceHost,
			appRegistryPathPrefix,
			appRegistryHost)

		req, err := http.NewRequest(http.MethodGet, "/path", nil)
		require.NoError(t, err)
		req.Header.Set(CertificateInfoHeader, certInfoHeader)
		recorder := httptest.NewRecorder()

		// when
		proxyHandler.ProxyAppConnectorRequests(recorder, req)

		// then
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("should return 404 when path is invalid", func(t *testing.T) {
		// given
		certInfoHeader :=
			`Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`

		proxyHandler := NewProxyHandler(
			"",
			"",
			eventServicePathPrefixV1,
			eventServicePathPrefixV2,
			eventServiceHost,
			appRegistryPathPrefix,
			appRegistryHost)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/v1/bad/path", applicationName), nil)
		require.NoError(t, err)
		req.Header.Set(CertificateInfoHeader, certInfoHeader)
		req = mux.SetURLVars(req, map[string]string{"application": applicationName})
		recorder := httptest.NewRecorder()

		// when
		proxyHandler.ProxyAppConnectorRequests(recorder, req)

		// then
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})

}
