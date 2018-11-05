package proxy

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authentication"
	authMock "github.com/kyma-project/kyma/components/proxy-service/internal/authentication/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httperrors"
	metadataMock "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxy(t *testing.T) {

	proxyTimeout := 10

	t.Run("should proxy", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "http://re-test-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("Setup", req).Return(nil)

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", authentication.Credentials{}).Return()

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
		}, nil).Times(1)

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
	})

	t.Run("should proxy OAuth calls", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		tsOAuth := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodPost)
			assert.Equal(t, req.RequestURI, "/token")
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "http://re-test-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("Setup", req).Return(nil)

		authCredentials := authentication.Credentials{
			Oauth: &authentication.OauthCredentials{
				ClientId:          "clientId",
				ClientSecret:      "clientSecret",
				AuthenticationUrl: tsOAuth.URL + "/token",
			},
		}

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", authCredentials).Return(authStrategyMock)

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
			Credentials: &serviceapi.Credentials{
				Oauth: serviceapi.Oauth{
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}, nil)

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
	})

	t.Run("should fail with Bad Gateway error when failed to get OAuth token", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		tsOAuth := NewForbiddenServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodPost)
			assert.Equal(t, req.RequestURI, "/token")
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "http://re-test-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("Setup", req).Return(nil)

		authCredentials := authentication.Credentials{
			Oauth: &authentication.OauthCredentials{
				ClientId:          "clientId",
				ClientSecret:      "clientSecret",
				AuthenticationUrl: tsOAuth.URL + "/token",
			},
		}

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", authCredentials).Return(authStrategyMock)

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
			Credentials: &serviceapi.Credentials{
				Oauth: serviceapi.Oauth{
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
		}, nil)

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadGateway, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
	})

	t.Run("should return 500 if failed to get service definition", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req.Host = "http://re-test-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local"
		rr := httptest.NewRecorder()

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").
			Return(&serviceapi.API{}, apperrors.Internal("Failed to read services"))

		handler := New(serviceDefServiceMock, nil, createProxyConfig(proxyTimeout))

		// when
		handler.ServeHTTP(rr, req)

		// then
		var errorResponse httperrors.ErrorResponse

		json.Unmarshal([]byte(rr.Body.String()), &errorResponse)

		serviceDefServiceMock.AssertExpectations(t)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
	})

	t.Run("should invalidate proxy and retry when 403 occurred", func(t *testing.T) {
		// given
		tsf := NewTestServerForRetryTest(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer tsf.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "http://re-test-c687e68a-9038-4f38-845b-9c61592e59e6.namespace.svc.cluster.local"

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: tsf.URL,
		}, nil)

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("Setup", req).Return(nil).Twice()
		authStrategyMock.On("Reset", req).Return().Once()

		authCredentials := authentication.Credentials{
			Basic: &authentication.BasicAuthCredentials{
				UserName: "username",
				Password: "password",
			},
		}

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", authCredentials).Return(authStrategyMock)

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		serviceDefServiceMock.AssertExpectations(t)
	})
}

func TestInvalidStateHandler(t *testing.T) {
	t.Run("should always return Internal Server Error", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler := NewInvalidStateHandler("Proxy Service id not initialized properly")

		// when
		handler.ServeHTTP(rr, req)

		// then
		var errorResponse httperrors.ErrorResponse

		json.Unmarshal([]byte(rr.Body.String()), &errorResponse)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
	})
}

func NewTestServer(check func(req *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("test"))
	}))
}

func NewForbiddenServer(check func(req *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
}

func NewTestServerForRetryTest(check func(req *http.Request)) *httptest.Server {
	var requestNotSent *bool

	*requestNotSent = false

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		if *requestNotSent {
			w.WriteHeader(http.StatusForbidden)
		} else {
			*requestNotSent = true
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}
	}))
}

func createProxyConfig(proxyTimeout int) Config {
	return Config{
		SkipVerify:        true,
		ProxyTimeout:      proxyTimeout,
		Namespace:         "kyma-integration",
		RemoteEnvironment: "test",
		ProxyCacheTTL:     proxyTimeout,
	}
}
