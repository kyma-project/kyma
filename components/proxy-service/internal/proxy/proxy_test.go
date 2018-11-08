package proxy

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization"
	authMock "github.com/kyma-project/kyma/components/proxy-service/internal/authorization/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httperrors"
	metadataMock "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxy(t *testing.T) {

	proxyTimeout := 10

	t.Run("should proxy and use internal cache", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "re-test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(nil).Twice()

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", authorization.Credentials{}).Return(authStrategyMock).Once()

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
		}, nil).Once()

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		// given
		nextReq, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		nextReq.Host = "re-test-uuid-1.namespace.svc.cluster.local"
		rr = httptest.NewRecorder()

		//when
		handler.ServeHTTP(rr, nextReq)

		//then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
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
		req.Host = "re-test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(nil)

		credentialsMatcher := createOAuthCredentialsMatcher("clientId", "clientSecret", tsOAuth.URL+"/token")

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.MatchedBy(credentialsMatcher)).Return(authStrategyMock)

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
			Credentials: &serviceapi.Credentials{
				Oauth: &serviceapi.Oauth{
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
					URL:          tsOAuth.URL + "/token",
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

	t.Run("should proxy Basic auth calls", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "re-test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(nil)

		credentialsMatcher := createBasicCredentialsMatcher("username", "password")

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.MatchedBy(credentialsMatcher)).Return(authStrategyMock)

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
			Credentials: &serviceapi.Credentials{
				Basic: &serviceapi.Basic{
					Username: "username",
					Password: "password",
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

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "re-test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.AnythingOfType("*http.Request")).Return(apperrors.UpstreamServerCallFailed("failed"))

		credentialsMatcher := createOAuthCredentialsMatcher("clientId", "clientSecret", "www.example.com/token")

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.MatchedBy(credentialsMatcher)).Return(authStrategyMock)

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
			Credentials: &serviceapi.Credentials{
				Oauth: &serviceapi.Oauth{
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
					URL:          "www.example.com/token",
				},
			},
		}, nil)

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadGateway, rr.Code)
	})

	t.Run("should return 500 if failed to get service definition", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req.Host = "re-test-uuid-1.namespace.svc.cluster.local"
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

	t.Run("should invalidate proxy and retry when 401 occurred", func(t *testing.T) {
		// given
		tsf := NewTestServerForRetryTest(http.StatusUnauthorized, func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer tsf.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "re-test-uuid-1.namespace.svc.cluster.local"

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: tsf.URL,
		}, nil).Twice()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.Anything).Return(nil).Twice()
		authStrategyMock.On("Invalidate").Return().Once()

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.Anything).Return(authStrategyMock).Twice()

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		serviceDefServiceMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
	})

	t.Run("should invalidate proxy and retry when 403 occurred", func(t *testing.T) {
		// given
		tsf := NewTestServerForRetryTest(http.StatusForbidden, func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer tsf.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "re-test-uuid-1.namespace.svc.cluster.local"

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: tsf.URL,
		}, nil).Twice()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.On("AddAuthorizationHeader", mock.Anything).Return(nil).Twice()
		authStrategyMock.On("Invalidate").Return().Once()

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.Anything).Return(authStrategyMock).Twice()

		handler := New(serviceDefServiceMock, authStrategyFactoryMock, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		serviceDefServiceMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
}

func NewTestServerForRetryTest(status int, check func(req *http.Request)) *httptest.Server {
	firstCall := true

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		if firstCall {
			w.WriteHeader(status)
			firstCall = false
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}
	}))
}

func createProxyConfig(proxyTimeout int) Config {
	return Config{
		SkipVerify:        true,
		ProxyTimeout:      proxyTimeout,
		RemoteEnvironment: "test",
		ProxyCacheTTL:     proxyTimeout,
	}
}

func createOAuthCredentialsMatcher(clientId, clientSecret, url string) func(authorization.Credentials) bool {
	return func(c authorization.Credentials) bool {
		return c.Oauth != nil && c.Oauth.ClientId == clientId &&
			c.Oauth.ClientSecret == clientSecret &&
			c.Oauth.Url == url
	}
}

func createBasicCredentialsMatcher(username, password string) func(authorization.Credentials) bool {
	return func(c authorization.Credentials) bool {
		return c.Basic != nil && c.Basic.Username == username &&
			c.Basic.Password == password
	}
}
