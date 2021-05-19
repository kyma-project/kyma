package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata"
	metadatamocks "github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	metadatamodel "github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	authMock "github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type createHandlerFunc func(serviceDefService metadata.ServiceDefinitionService, authorizationStrategyFactory authorization.StrategyFactory, csrfTokenStrategyFactory csrf.TokenStrategyFactory, config Config) http.Handler

func TestProxyFactory(t *testing.T) {

	type createMockServiceDefServiceFunc func(apiIdentifier model.APIIdentifier, targetURL string, credentials *authorization.Credentials) metadatamocks.ServiceDefinitionService

	type testcase struct {
		name                            string
		url                             string
		expectedTargetAPIUrl            string
		createHandlerFunc               createHandlerFunc
		createMockServiceDefServiceFunc createMockServiceDefServiceFunc
		apiIdentifier                   metadatamodel.APIIdentifier
	}

	proxyConfig := Config{
		SkipVerify:    true,
		ProxyTimeout:  10,
		Application:   "test",
		ProxyCacheTTL: 10,
	}

	createTestServer := func(path string) *httptest.Server {
		return NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, path)
		})
	}

	createMockServiceDeffService := func(apiIdentifier model.APIIdentifier, targetURL string, credentials *authorization.Credentials) metadatamocks.ServiceDefinitionService {
		serviceDefServiceMock := metadatamocks.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPIByServiceName", apiIdentifier.Application, apiIdentifier.Service).Return(&metadatamodel.API{
			TargetUrl:   targetURL,
			Credentials: credentials,
		}, nil).Once()

		return serviceDefServiceMock
	}

	createMockServiceDeffServiceForCompass := func(apiIdentifier model.APIIdentifier, targetURL string, credentials *authorization.Credentials) metadatamocks.ServiceDefinitionService {
		serviceDefServiceMock := metadatamocks.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPIByEntryName", apiIdentifier.Application, apiIdentifier.Service, apiIdentifier.Entry).Return(&metadatamodel.API{
			TargetUrl:   targetURL,
			Credentials: credentials,
		}, nil).Once()

		return serviceDefServiceMock
	}

	apiIdentifier := metadatamodel.APIIdentifier{
		Application: "app",
		Service:     "service",
	}

	apiIdentifierForCompass := metadatamodel.APIIdentifier{
		Application: "app",
		Service:     "service",
		Entry:       "entry",
	}

	for _, testCase := range []testcase{
		{
			name:                            "should proxy using application and service name",
			url:                             "/app/service/orders/123",
			expectedTargetAPIUrl:            "/orders/123",
			createHandlerFunc:               New,
			createMockServiceDefServiceFunc: createMockServiceDeffService,
			apiIdentifier:                   apiIdentifier,
		},
		{
			name:                            "should proxy using application and service name when accessing root path",
			url:                             "/app/service",
			expectedTargetAPIUrl:            "/",
			createHandlerFunc:               New,
			createMockServiceDefServiceFunc: createMockServiceDeffService,
			apiIdentifier:                   apiIdentifier,
		},
		{
			name:                            "should proxy using application, service and entry name",
			url:                             "/app/service/entry/orders/123",
			expectedTargetAPIUrl:            "/orders/123",
			createHandlerFunc:               NewForCompass,
			createMockServiceDefServiceFunc: createMockServiceDeffServiceForCompass,
			apiIdentifier:                   apiIdentifierForCompass,
		},
		{
			name:                            "should proxy using application, service and entry name when accessing root path",
			url:                             "/app/service/entry",
			expectedTargetAPIUrl:            "/",
			createHandlerFunc:               NewForCompass,
			createMockServiceDefServiceFunc: createMockServiceDeffServiceForCompass,
			apiIdentifier:                   apiIdentifierForCompass,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			// given
			ts := createTestServer(testCase.expectedTargetAPIUrl)
			req, err := http.NewRequest(http.MethodGet, testCase.url, nil)
			require.NoError(t, err)

			authStrategyMock := &authMock.Strategy{}
			authStrategyMock.
				On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
				Return(nil).
				Once()

			credentials := &authorization.Credentials{}
			authStrategyFactoryMock := &authMock.StrategyFactory{}
			authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

			csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

			serviceDefServiceMock := testCase.createMockServiceDefServiceFunc(testCase.apiIdentifier, ts.URL, credentials)

			handler := testCase.createHandlerFunc(&serviceDefServiceMock, authStrategyFactoryMock, csrfFactoryMock, proxyConfig)
			rr := httptest.NewRecorder()

			// when
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "test", rr.Body.String())
			serviceDefServiceMock.AssertExpectations(t)
			authStrategyFactoryMock.AssertExpectations(t)
			authStrategyMock.AssertExpectations(t)
			csrfFactoryMock.AssertExpectations(t)
			csrfStrategyMock.AssertExpectations(t)
		})
	}
}

func TestPathExtractionErrors(t *testing.T) {
	proxyConfig := Config{
		SkipVerify:    true,
		ProxyTimeout:  10,
		Application:   "test",
		ProxyCacheTTL: 10,
	}

	type testcase struct {
		name              string
		url               string
		createHandlerFunc createHandlerFunc
	}

	testCases := []testcase{
		{
			name:              "Should return Internal error when failed to extract data from empty path",
			url:               "",
			createHandlerFunc: New,
		},
		{
			name:              "Should return Internal error when failed to extract data from path containing application name only",
			url:               "/appName",
			createHandlerFunc: New,
		},
		{
			name:              "Should return Internal error when failed to extract data from empty path (Compass)",
			url:               "",
			createHandlerFunc: NewForCompass,
		},
		{
			name:              "Should return Internal error when failed to extract data from path containing application name only (Compass)",
			url:               "/appName",
			createHandlerFunc: NewForCompass,
		},
		{
			name:              "Should return Internal error when failed to extract data from path containing application and service name only (Compass)",
			url:               "/appName/serviceName",
			createHandlerFunc: NewForCompass,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// given
			handler := testCase.createHandlerFunc(nil, nil, nil, proxyConfig)

			req, err := http.NewRequest(http.MethodGet, testCase.url, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()

			// when
			handler.ServeHTTP(rr, req)

			// then
			assert.Equal(t, http.StatusInternalServerError, rr.Code)
		})
	}
}
