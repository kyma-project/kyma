package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httperrors"
	k8smocks "github.com/kyma-project/kyma/components/proxy-service/internal/k8sconsts/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata"
	metadataMock "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/proxy-service/internal/proxy/mocks"
	"github.com/kyma-project/kyma/components/proxy-service/internal/proxy/proxycache"
	cacheMock "github.com/kyma-project/kyma/components/proxy-service/internal/proxy/proxycache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
		req.Host = "uuid-1.cluster.local"

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return("uuid-1")

		u, _ := url.Parse(ts.URL)
		httpCacheMock := &cacheMock.HTTPProxyCache{}
		httpCacheMock.On("Get", "uuid-1").Return(
			&proxycache.Proxy{
				Proxy:        httputil.NewSingleHostReverseProxy(u),
				ClientId:     "",
				OauthUrl:     "",
				ClientSecret: "",
			}, true)

		handler := New(nameResolver, nil, nil, httpCacheMock, true, proxyTimeout)
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		httpCacheMock.AssertExpectations(t)
	})

	t.Run("should proxy with prefetching oauth token", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "Bearer access_token", req.Header.Get(httpconsts.HeaderAuthorization))
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "uuid-1.cluster.local"

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return("uuid-1")

		oauthClientMock := &mocks.OAuthClient{}
		oauthClientMock.On(
			"GetToken",
			"clientId",
			"clientSecret",
			"www.example.com/oauth",
		).Return("Bearer access_token", nil)

		u, _ := url.Parse(ts.URL)
		httpCacheMock := &cacheMock.HTTPProxyCache{}
		httpCacheMock.On("Get", "uuid-1").Return(
			&proxycache.Proxy{
				Proxy:        httputil.NewSingleHostReverseProxy(u),
				ClientId:     "clientId",
				ClientSecret: "clientSecret",
				OauthUrl:     "www.example.com/oauth",
			}, true)

		handler := New(nameResolver, nil, oauthClientMock, httpCacheMock, true, proxyTimeout)
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		oauthClientMock.AssertExpectations(t)
		httpCacheMock.AssertExpectations(t)
	})

	t.Run("should handle Kyma-Target-Token header", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "token", req.Header.Get(httpconsts.HeaderAuthorization))
			assert.Equal(t, "", req.Header.Get(httpconsts.HeaderAccessToken))
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "uuid-1.cluster.local"
		req.Header.Set(httpconsts.HeaderAccessToken, "token")

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return("uuid-1")

		u, _ := url.Parse(ts.URL)
		httpCacheMock := &cacheMock.HTTPProxyCache{}
		httpCacheMock.On("Get", "uuid-1").Return(
			&proxycache.Proxy{
				Proxy:        httputil.NewSingleHostReverseProxy(u),
				ClientId:     "clientId",
				ClientSecret: "clientSecret",
				OauthUrl:     "www.example.com/oauth",
			}, true)

		handler := New(nameResolver, nil, nil, httpCacheMock, true, proxyTimeout)
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		httpCacheMock.AssertExpectations(t)
	})

	t.Run("should proxy on cache miss", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
			assert.Equal(t, "", req.Header.Get(httpconsts.HeaderAuthorization))
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "uuid-1.cluster.local"

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return("uuid-1")

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: ts.URL,
		}, nil)

		u, _ := url.Parse(ts.URL)
		httpCacheMock := &cacheMock.HTTPProxyCache{}
		httpCacheMock.On("Get", "uuid-1").Return(nil, false)
		httpCacheMock.On("Add", "uuid-1", "", "", "", mock.AnythingOfType("*httputil.ReverseProxy")).Return(
			&proxycache.Proxy{
				Proxy:        httputil.NewSingleHostReverseProxy(u),
				ClientId:     "",
				OauthUrl:     "",
				ClientSecret: "",
			})

		handler := New(nameResolver, serviceDefServiceMock, nil, httpCacheMock, true, proxyTimeout)
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		serviceDefServiceMock.AssertExpectations(t)
		httpCacheMock.AssertExpectations(t)
	})

	t.Run("should proxy on cache miss with prefetching oauth token ", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "Bearer access_token", req.Header.Get(httpconsts.HeaderAuthorization))
		})
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "uuid-1.cluster.local"

		serviceDefinition := metadata.ServiceDefinition{
			ID:   "uuid-1",
			Name: "service1",
			Api: &serviceapi.API{
				TargetUrl: ts.URL,
				Credentials: &serviceapi.Credentials{
					Oauth: serviceapi.Oauth{
						URL:          "www.example.com/oauth",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
			}}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return(serviceDefinition.ID)

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", serviceDefinition.ID).Return(serviceDefinition.Api, nil)

		oauthClientMock := &mocks.OAuthClient{}
		oauthClientMock.On(
			"GetToken",
			serviceDefinition.Api.Credentials.Oauth.ClientID,
			serviceDefinition.Api.Credentials.Oauth.ClientSecret,
			serviceDefinition.Api.Credentials.Oauth.URL,
		).Return("Bearer access_token", nil)

		u, _ := url.Parse(serviceDefinition.Api.TargetUrl)
		httpCacheMock := &cacheMock.HTTPProxyCache{}
		httpCacheMock.On("Get", "uuid-1").Return(nil, false)
		httpCacheMock.On(
			"Add",
			"uuid-1",
			serviceDefinition.Api.Credentials.Oauth.URL,
			serviceDefinition.Api.Credentials.Oauth.ClientID,
			serviceDefinition.Api.Credentials.Oauth.ClientSecret,
			mock.AnythingOfType("*httputil.ReverseProxy"),
		).Return(
			&proxycache.Proxy{
				Proxy:        httputil.NewSingleHostReverseProxy(u),
				ClientId:     serviceDefinition.Api.Credentials.Oauth.ClientID,
				OauthUrl:     serviceDefinition.Api.Credentials.Oauth.URL,
				ClientSecret: serviceDefinition.Api.Credentials.Oauth.ClientSecret,
			})

		handler := New(nameResolver, serviceDefServiceMock, oauthClientMock, httpCacheMock, true, proxyTimeout)
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		serviceDefServiceMock.AssertExpectations(t)
		oauthClientMock.AssertExpectations(t)
		httpCacheMock.AssertExpectations(t)
	})

	t.Run("should return 500 if failed to get service definition", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req.Host = "uuid-1.cluster.local"
		rr := httptest.NewRecorder()

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return("uuid-1")

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").
			Return(&serviceapi.API{}, apperrors.Internal("Failed to read services"))

		proxyCacheMock := &cacheMock.HTTPProxyCache{}
		proxyCacheMock.On("Get", "uuid-1").Return(nil, false)

		handler := New(nameResolver, serviceDefServiceMock, nil, proxyCacheMock, true, proxyTimeout)

		// when
		handler.ServeHTTP(rr, req)

		// then
		var errorResponse httperrors.ErrorResponse

		json.Unmarshal([]byte(rr.Body.String()), &errorResponse)

		serviceDefServiceMock.AssertExpectations(t)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)

		proxyCacheMock.AssertExpectations(t)
	})

	t.Run("should return 502 if failed to prefetch token", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req.Host = "uuid-1.cluster.local"
		rr := httptest.NewRecorder()

		serviceDefinition := metadata.ServiceDefinition{
			ID:   "uuid-1",
			Name: "service1",
			Api: &serviceapi.API{
				TargetUrl: "www.exaple.com/service1",
				Credentials: &serviceapi.Credentials{
					Oauth: serviceapi.Oauth{
						URL:          "www.example.com/oauth",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
			}}

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return("uuid-1")

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}

		oauthClientMock := &mocks.OAuthClient{}
		oauthClientMock.On(
			"GetToken",
			serviceDefinition.Api.Credentials.Oauth.ClientID,
			serviceDefinition.Api.Credentials.Oauth.ClientSecret,
			serviceDefinition.Api.Credentials.Oauth.URL,
		).Return("", apperrors.UpstreamServerCallFailed("failed to get token"))

		httpCacheMock := &cacheMock.HTTPProxyCache{}
		httpCacheMock.On("Get", "uuid-1").
			Return(&proxycache.Proxy{
				Proxy:        &httputil.ReverseProxy{},
				OauthUrl:     serviceDefinition.Api.Credentials.Oauth.URL,
				ClientId:     serviceDefinition.Api.Credentials.Oauth.ClientID,
				ClientSecret: serviceDefinition.Api.Credentials.Oauth.ClientSecret,
			}, true)

		handler := New(nameResolver, serviceDefServiceMock, oauthClientMock, httpCacheMock, true, 10)

		// when
		handler.ServeHTTP(rr, req)

		// then
		var errorResponse httperrors.ErrorResponse

		json.Unmarshal([]byte(rr.Body.String()), &errorResponse)

		assert.Equal(t, http.StatusBadGateway, rr.Code)
		assert.Equal(t, http.StatusBadGateway, errorResponse.Code)

		serviceDefServiceMock.AssertExpectations(t)
		httpCacheMock.AssertExpectations(t)
		oauthClientMock.AssertExpectations(t)
	})

	t.Run("should invalidate proxy and retry when 403 occurred", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		tsf := NewForbiddenTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer tsf.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		req.Host = "uuid-1.cluster.local"

		nameResolver := new(k8smocks.NameResolver)
		nameResolver.On("ExtractServiceId", "uuid-1.cluster.local").Return("uuid-1")

		serviceDefServiceMock := &metadataMock.ServiceDefinitionService{}
		serviceDefServiceMock.On("GetAPI", "uuid-1").Return(&serviceapi.API{
			TargetUrl: tsf.URL,
		}, nil)

		u, _ := url.Parse(ts.URL)
		httpCacheMock := &cacheMock.HTTPProxyCache{}
		httpCacheMock.On("Get", "uuid-1").Return(nil, false)
		httpCacheMock.On("Add", "uuid-1", "", "", "", mock.AnythingOfType("*httputil.ReverseProxy")).Return(
			&proxycache.Proxy{
				Proxy:        httputil.NewSingleHostReverseProxy(u),
				ClientId:     "",
				OauthUrl:     "",
				ClientSecret: "",
			})

		handler := New(nameResolver, serviceDefServiceMock, nil, httpCacheMock, true, proxyTimeout)
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		serviceDefServiceMock.AssertExpectations(t)
		httpCacheMock.AssertExpectations(t)
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

func NewForbiddenTestServer(check func(req *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("test"))
	}))
}
