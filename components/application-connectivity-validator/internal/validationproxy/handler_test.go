package validationproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appconnv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

const (
	applicationName     = "test-application"
	applicationMetaName = "test-application-meta"
	applicationID       = "test-application-id"

	group  = "group"
	tenant = "tenant"

	eventingPathPrefixV1           = "/test-application/v1/events"
	eventingPathPrefixV2           = "/test-application/v2/events"
	eventingPathPrefixEvents       = "/test-application/events"
	appRegistryPathPrefix          = "/test-application/v1/metadata"
	eventingDestinationPathPublish = "/publish"
	eventingDestinationPath        = "/"
)

type event struct {
	Title string `json:"title"`
}

var (
	applicationManagedByCompass = &appconnv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationMetaName,
		},
		Spec: appconnv1alpha1.ApplicationSpec{
			Description:     "Description",
			Services:        []appconnv1alpha1.Service{},
			CompassMetadata: &appconnv1alpha1.CompassMetadata{Authentication: appconnv1alpha1.Authentication{ClientIds: []string{applicationID}}},
		},
	}
	applicationNotManagedByCompass = &appconnv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
		},
		Spec: appconnv1alpha1.ApplicationSpec{
			Description: "Description",
			Services:    []appconnv1alpha1.Service{},
		},
	}
)

func TestProxyHandler_ProxyAppConnectorRequests(t *testing.T) {
	log, err := logger.New(logger.TEXT, logger.ERROR)
	require.NoError(t, err)
	testCases := []struct {
		caseDescription string
		tenant          string
		group           string
		certInfoHeader  string
		expectedStatus  int
		application     *appconnv1alpha1.Application
	}{
		{
			caseDescription: "Application without group and tenant",
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application-id,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusOK,
			application:    applicationManagedByCompass,
		},
		{
			caseDescription: "Application without group and tenant and with invalid Common Name",
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=invalid-cn,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
			application:    applicationManagedByCompass,
		},
		{
			caseDescription: "Application with group and tenant",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application-id,OU=group,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusOK,
			application:    applicationManagedByCompass,
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
			application:    applicationManagedByCompass,
		},
		{
			caseDescription: "Application with group, tenant and invalid Organizational Unit",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application-id,OU=invalid,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
			application:    applicationManagedByCompass,
		},
		{
			caseDescription: "Application with group, tenant and invalid Organization",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application-id,OU=group,O=invalid,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
			application:    applicationManagedByCompass,
		},
		{
			caseDescription: "X-Forwarded-Client-Cert header not specified",
			tenant:          tenant,
			group:           group,
			expectedStatus:  http.StatusInternalServerError,
			application:     applicationManagedByCompass,
		},
		{
			caseDescription: "Application not managed by Compass Runtime Agent without group and tenant",
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusOK,
			application:    applicationNotManagedByCompass,
		},
		{
			caseDescription: "Application not managed by Compass Runtime Agent without group and tenant and with invalid Common Name",
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=invalid-cn,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
			application:    applicationNotManagedByCompass,
		},
		{
			caseDescription: "Application not managed by Compass Runtime Agent with group and tenant",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=group,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusOK,
			application:    applicationNotManagedByCompass,
		},
		{
			caseDescription: "Application not managed by Compass Runtime Agent with group, tenant and invalid Common Name",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=invalid-application,OU=group,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
			application:    applicationNotManagedByCompass,
		},
		{
			caseDescription: "Application not managed by Compass Runtime Agent with group, tenant and invalid Organizational Unit",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=invalid,O=tenant,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
			application:    applicationNotManagedByCompass,
		},
		{
			caseDescription: "Application not managed by Compass Runtime Agent with group, tenant and invalid Organization",
			tenant:          tenant,
			group:           group,
			certInfoHeader: `Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=group,O=invalid,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`,
			expectedStatus: http.StatusForbidden,
			application:    applicationNotManagedByCompass,
		},
	}

	t.Run("should proxy requests", func(t *testing.T) {
		eventingPublisherHandler := mux.NewRouter()
		eventingPublisherServer := httptest.NewServer(eventingPublisherHandler)
		eventingPublisherHost := strings.TrimPrefix(eventingPublisherServer.URL, "http://")

		appRegistryHandler := mux.NewRouter()
		appRegistryServer := httptest.NewServer(appRegistryHandler)
		appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")

		for _, testCase := range testCases {
			// given
			idCache := cache.New(time.Minute, time.Minute)
			if testCase.application.Spec.CompassMetadata != nil {
				idCache.Set(testCase.application.Name, []string{applicationID}, cache.NoExpiration)
			} else {
				idCache.Set(testCase.application.Name, []string{}, cache.NoExpiration)
			}

			eventingPathPrefixV1 := fmt.Sprintf("/%s/v1/events", testCase.application.Name)
			eventingPathPrefixV2 := fmt.Sprintf("/%s/v2/events", testCase.application.Name)
			eventingPathPrefixEvents := fmt.Sprintf("/%s/events", testCase.application.Name)
			appRegistryPathPrefix := fmt.Sprintf("/%s/v1/metadata", testCase.application.Name)
			proxyHandler := NewProxyHandler(
				testCase.group,
				testCase.tenant,
				eventingPathPrefixV1,
				eventingPathPrefixV2,
				eventingPublisherHost,
				eventingPathPrefixEvents,
				eventingDestinationPath,
				appRegistryPathPrefix,
				appRegistryHost,
				idCache,
				log)

			t.Run("should proxy eventing V1 request when "+testCase.caseDescription, func(t *testing.T) {
				eventTitle := "my-event-1"

				eventingPublisherHandler.PathPrefix("/{application}/v1/events").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					appName := mux.Vars(r)["application"]
					assert.Equal(t, testCase.application.Name, appName, `Error reading "application" route variable from request context`)

					var receivedEvent event

					err := json.NewDecoder(r.Body).Decode(&receivedEvent)
					require.NoError(t, err)
					assert.Equal(t, eventTitle, receivedEvent.Title)

					w.WriteHeader(http.StatusOK)
				})

				body, err := json.Marshal(event{Title: eventTitle})
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/v1/events", testCase.application.Name), bytes.NewReader(body))
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})

				recorder := httptest.NewRecorder()

				// when
				proxyHandler.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})

			t.Run("should proxy eventing request when "+testCase.caseDescription, func(t *testing.T) {
				const eventTitle = "my-event"
				const mockIncomingRequestHost = "fake.istio.gateway"

				eventingPublisherHandler.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var receivedEvent event

					err := json.NewDecoder(r.Body).Decode(&receivedEvent)
					require.NoError(t, err)
					assert.Equal(t, eventTitle, receivedEvent.Title)

					assert.NotEqual(t, mockIncomingRequestHost, r.Host, "proxy should rewrite Host field")

					w.WriteHeader(http.StatusOK)
				})

				body, err := json.Marshal(event{Title: eventTitle})
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/events", testCase.application.Name), bytes.NewReader(body))
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})

				// mock request Host to assert it gets rewritten by the proxy
				req.Host = mockIncomingRequestHost

				recorder := httptest.NewRecorder()

				// when
				proxyHandler.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})

			t.Run("should proxy application registry request when "+testCase.caseDescription, func(t *testing.T) {
				appRegistryHandler.PathPrefix("/{application}/v1/metadata/services").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					appName := mux.Vars(r)["application"]
					assert.Equal(t, testCase.application.Name, appName, `Error reading "application" route variable from request context`)
					w.WriteHeader(http.StatusOK)
				})

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/v1/metadata/services", testCase.application.Name), nil)
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})

				recorder := httptest.NewRecorder()

				// when
				proxyHandler.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})
		}
	})

	t.Run("should use cached IDs while proxying requests", func(t *testing.T) {
		eventingPublisherHandler := mux.NewRouter()
		eventingPublisherServer := httptest.NewServer(eventingPublisherHandler)
		eventingPublisherHost := strings.TrimPrefix(eventingPublisherServer.URL, "http://")

		appRegistryHandler := mux.NewRouter()
		appRegistryServer := httptest.NewServer(appRegistryHandler)
		appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")

		for _, testCase := range testCases {
			// given
			idCache := cache.New(time.Minute, time.Minute)
			if testCase.application.Spec.CompassMetadata != nil {
				idCache.Set(testCase.application.Name, []string{applicationID}, cache.NoExpiration)
			} else {
				idCache.Set(testCase.application.Name, []string{}, cache.NoExpiration)
			}

			eventingPathPrefixV1 := fmt.Sprintf("/%s/v1/events", testCase.application.Name)
			eventingPathPrefixV2 := fmt.Sprintf("/%s/v2/events", testCase.application.Name)
			eventingPathPrefixEvents := fmt.Sprintf("/%s/events", testCase.application.Name)
			appRegistryPathPrefix := fmt.Sprintf("/%s/v1/metadata", testCase.application.Name)
			proxyHandler := NewProxyHandler(
				testCase.group,
				testCase.tenant,
				eventingPathPrefixV1,
				eventingPathPrefixV2,
				eventingPublisherHost,
				eventingPathPrefixEvents,
				eventingDestinationPath,
				appRegistryPathPrefix,
				appRegistryHost,
				idCache,
				log)

			t.Run("should proxy application registry request when "+testCase.caseDescription, func(t *testing.T) {
				appRegistryHandler.PathPrefix("/{application}/v1/metadata/services").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					appName := mux.Vars(r)["application"]
					assert.Equal(t, testCase.application.Name, appName)
					w.WriteHeader(http.StatusOK)
				})

				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/v1/metadata/services", testCase.application.Name), nil)
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})
				recorder := httptest.NewRecorder()

				// when
				proxyHandler.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})
		}
	})

	t.Run("should return 500 failed when cache doesn't contain the element", func(t *testing.T) {
		eventingPublisherHandler := mux.NewRouter()
		eventingPublisherServer := httptest.NewServer(eventingPublisherHandler)
		eventingPublisherHost := strings.TrimPrefix(eventingPublisherServer.URL, "http://")

		appRegistryHandler := mux.NewRouter()
		appRegistryServer := httptest.NewServer(appRegistryHandler)
		appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")

		for _, testCase := range testCases {
			// given
			idCache := cache.New(time.Minute, time.Minute)

			eventingPathPrefixV1 := fmt.Sprintf("/%s/v1/events", testCase.application.Name)
			eventingPathPrefixV2 := fmt.Sprintf("/%s/v2/events", testCase.application.Name)
			eventingPathPrefixEvents := fmt.Sprintf("/%s/events", testCase.application.Name)
			appRegistryPathPrefix := fmt.Sprintf("/%s/v1/metadata", testCase.application.Name)
			proxyHandler := NewProxyHandler(
				testCase.group,
				testCase.tenant,
				eventingPathPrefixV1,
				eventingPathPrefixV2,
				eventingPublisherHost,
				eventingPathPrefixEvents,
				eventingDestinationPath,
				appRegistryPathPrefix,
				appRegistryHost,
				idCache,
				log)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/v1/metadata/services", testCase.application.Name), nil)
			require.NoError(t, err)
			req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
			req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})
			recorder := httptest.NewRecorder()

			// when
			proxyHandler.ProxyAppConnectorRequests(recorder, req)

			// then
			assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		}
	})

	t.Run("should return 400 when application not specified in path", func(t *testing.T) {
		eventingPublisherHandler := mux.NewRouter()
		eventingPublisherServer := httptest.NewServer(eventingPublisherHandler)
		eventingPublisherHost := strings.TrimPrefix(eventingPublisherServer.URL, "http://")

		appRegistryHandler := mux.NewRouter()
		appRegistryServer := httptest.NewServer(appRegistryHandler)
		appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")
		// given
		certInfoHeader :=
			`Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";URI=`

		idCache := cache.New(time.Minute, time.Minute)
		idCache.Set(applicationName, []string{}, cache.NoExpiration)

		proxyHandler := NewProxyHandler(
			group,
			tenant,
			eventingPathPrefixV1,
			eventingPathPrefixV2,
			eventingPublisherHost,
			eventingPathPrefixEvents,
			eventingDestinationPath,
			appRegistryPathPrefix,
			appRegistryHost,
			idCache,
			log)

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
		eventingPublisherHandler := mux.NewRouter()
		eventingPublisherServer := httptest.NewServer(eventingPublisherHandler)
		eventingPublisherHost := strings.TrimPrefix(eventingPublisherServer.URL, "http://")

		appRegistryHandler := mux.NewRouter()
		appRegistryServer := httptest.NewServer(appRegistryHandler)
		appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")
		// given
		certInfoHeader :=
			`Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application-id,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";` +
				`URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;` +
				`Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";` +
				`URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`

		// mock cache sync controller that it fills cache
		idCache := cache.New(time.Minute, time.Minute)
		idCache.Set(applicationMetaName, []string{applicationID}, cache.NoExpiration)

		proxyHandler := NewProxyHandler(
			"",
			"",
			eventingPathPrefixV1,
			eventingPathPrefixV2,
			eventingPublisherHost,
			eventingPathPrefixEvents,
			eventingDestinationPath,
			appRegistryPathPrefix,
			appRegistryHost,
			idCache,
			log)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s/v1/bad/path", applicationMetaName), nil)
		require.NoError(t, err)
		req.Header.Set(CertificateInfoHeader, certInfoHeader)
		req = mux.SetURLVars(req, map[string]string{"application": applicationMetaName})
		recorder := httptest.NewRecorder()

		// when
		proxyHandler.ProxyAppConnectorRequests(recorder, req)

		// then
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})

	t.Run("should proxy requests to Event Publisher Proxy(EPP)", func(t *testing.T) {
		appRegistryHandler := mux.NewRouter()
		appRegistryServer := httptest.NewServer(appRegistryHandler)
		appRegistryHost := strings.TrimPrefix(appRegistryServer.URL, "http://")

		eventPublisherV1ProxyHandler := mux.NewRouter()
		eventPublisherV1ProxyServer := httptest.NewServer(eventPublisherV1ProxyHandler)
		eventPublisherV1ProxyHost := strings.TrimPrefix(eventPublisherV1ProxyServer.URL, "http://")

		for _, testCase := range testCases {
			// given
			idCache := cache.New(time.Minute, time.Minute)
			if testCase.application.Spec.CompassMetadata != nil {
				idCache.Set(testCase.application.Name, []string{applicationID}, cache.NoExpiration)
			} else {
				idCache.Set(testCase.application.Name, []string{}, cache.NoExpiration)
			}

			eventingPathPrefixV1 := fmt.Sprintf("/%s/v1/events", testCase.application.Name)
			eventingPathPrefixV2 := fmt.Sprintf("/%s/v2/events", testCase.application.Name)
			eventingPathPrefixEvents := fmt.Sprintf("/%s/events", testCase.application.Name)
			appRegistryPathPrefix := fmt.Sprintf("/%s/v1/metadata", testCase.application.Name)

			t.Run("should proxy requests in V1 to V1 endpoint of EPP when "+testCase.caseDescription, func(t *testing.T) {

				proxyHandlerBEB := NewProxyHandler(
					testCase.group,
					testCase.tenant,
					eventingPathPrefixV1,
					eventingPathPrefixV2,
					eventPublisherV1ProxyHost,
					eventingPathPrefixEvents,
					eventingDestinationPathPublish,
					appRegistryPathPrefix,
					appRegistryHost,
					idCache,
					log)
				eventTitle := "my-event-1"

				eventPublisherV1ProxyHandler.PathPrefix("/{application}/v1/events").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					appName := mux.Vars(r)["application"]
					assert.Equal(t, testCase.application.Name, appName, `Error reading "application" route variable from request context`)

					var receivedEvent event

					err := json.NewDecoder(r.Body).Decode(&receivedEvent)
					require.NoError(t, err)
					assert.Equal(t, eventTitle, receivedEvent.Title)

					w.WriteHeader(http.StatusOK)
				})

				body, err := json.Marshal(event{Title: eventTitle})
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/v1/events", testCase.application.Name), bytes.NewReader(body))
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})

				recorder := httptest.NewRecorder()

				// when
				proxyHandlerBEB.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})

			t.Run("should proxy requests in V2 to /publish endpoint of EPP when "+testCase.caseDescription, func(t *testing.T) {
				eventTitle := "my-event-2"

				eventPublisherProxyHandler := mux.NewRouter()
				eventPublisherProxyServer := httptest.NewServer(eventPublisherProxyHandler)
				eventPublisherProxyHost := strings.TrimPrefix(eventPublisherProxyServer.URL, "http://")

				proxyHandlerBEB := NewProxyHandler(
					testCase.group,
					testCase.tenant,
					eventingPathPrefixV1,
					eventingPathPrefixV2,
					eventPublisherProxyHost,
					eventingPathPrefixEvents,
					eventingDestinationPathPublish,
					appRegistryPathPrefix,
					appRegistryHost,
					idCache,
					log)

				eventPublisherProxyHandler.PathPrefix("/publish").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var receivedEvent event

					err := json.NewDecoder(r.Body).Decode(&receivedEvent)
					require.NoError(t, err)
					assert.Equal(t, eventTitle, receivedEvent.Title)

					w.WriteHeader(http.StatusOK)
				})

				body, err := json.Marshal(event{Title: eventTitle})
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/v2/events", testCase.application.Name), bytes.NewReader(body))
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})

				recorder := httptest.NewRecorder()

				// when
				proxyHandlerBEB.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})

			t.Run("should proxy requests in /events to /publish endpoint of EPP when "+testCase.caseDescription, func(t *testing.T) {
				const eventTitle = "my-event"
				const mockIncomingRequestHost = "fake.istio.gateway"

				eventPublisherProxyHandler := mux.NewRouter()
				eventPublisherProxyServer := httptest.NewServer(eventPublisherProxyHandler)
				eventPublisherProxyHost := strings.TrimPrefix(eventPublisherProxyServer.URL, "http://")
				proxyHandlerBEB := NewProxyHandler(
					testCase.group,
					testCase.tenant,
					eventingPathPrefixV1,
					eventingPathPrefixV2,
					eventPublisherProxyHost,
					eventingPathPrefixEvents,
					eventingDestinationPathPublish,
					appRegistryPathPrefix,
					appRegistryHost,
					idCache,
					log)

				eventPublisherProxyHandler.Path("/publish").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var receivedEvent event

					err := json.NewDecoder(r.Body).Decode(&receivedEvent)
					require.NoError(t, err)
					assert.Equal(t, eventTitle, receivedEvent.Title)

					assert.NotEqual(t, mockIncomingRequestHost, r.Host, "proxy should rewrite Host field")

					w.WriteHeader(http.StatusOK)
				})

				body, err := json.Marshal(event{Title: eventTitle})
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s/events", testCase.application.Name), bytes.NewReader(body))
				require.NoError(t, err)
				req.Header.Set(CertificateInfoHeader, testCase.certInfoHeader)
				req = mux.SetURLVars(req, map[string]string{"application": testCase.application.Name})

				// mock request Host to assert it gets rewritten by the proxy
				req.Host = mockIncomingRequestHost

				recorder := httptest.NewRecorder()

				// when
				proxyHandlerBEB.ProxyAppConnectorRequests(recorder, req)

				// then
				assert.Equal(t, testCase.expectedStatus, recorder.Code)
			})
		}
	})
}
