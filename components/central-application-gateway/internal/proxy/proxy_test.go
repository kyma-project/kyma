package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	csrfMock "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/httperrors"
	metadatamodel "github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	proxyMocks "github.com/kyma-project/kyma/components/central-application-gateway/internal/proxy/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	authMock "github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	secretName = "my-secret"
	apiName    = "my-api"
)

func TestProxy(t *testing.T) {

	proxyTimeout := 10
	apiIdentifier := metadatamodel.APIIdentifier{
		Application: "app",
		Service:     "service",
		Entry:       "entry",
	}

	fakePathExtractor := func(path string) (metadatamodel.APIIdentifier, string, apperrors.AppError) {

		apiIdentifier := metadatamodel.APIIdentifier{
			Application: "app",
			Service:     "service",
			Entry:       "entry",
		}

		return apiIdentifier, path, nil
	}

	t.Run("should proxy without escaping the URL path characters when target URL does not contain path", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "/somepath/Xyz('123')", req.URL.String())
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "somepath/Xyz('123')", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil).
			Once()

		credentials := &authorization.Credentials{}
		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:   ts.URL,
			Credentials: credentials,
		}, nil).Once()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should proxy without escaping the URL path characters when target URL contains path", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "/somepath/Xyz('123')", req.URL.String())
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "/Xyz('123')", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil).
			Once()

		credentials := &authorization.Credentials{}
		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:   ts.URL + "/somepath",
			Credentials: credentials,
		}, nil).Once()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should proxy without escaping the URL path characters when target URL contains full path", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "/somepath/Xyz('123')?$search=XXX", req.URL.String())
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "?$search=XXX", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil).
			Once()

		credentials := &authorization.Credentials{}
		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:   ts.URL + "/somepath/Xyz('123')",
			Credentials: credentials,
		}, nil).Once()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should proxy and add additional query parameters", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "param-value-1", req.URL.Query().Get("param1"))

			assert.Equal(t, 2, len(req.URL.Query()["param2"]))
			assert.Equal(t, "param-value-2.1", req.URL.Query().Get("param2"))
			assert.Equal(t, "param-value-2.1", req.URL.Query()["param2"][0])
			assert.Equal(t, "param-value-2.2", req.URL.Query()["param2"][1])
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil).
			Once()

		credentials := &authorization.Credentials{}
		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		requestParameters := &authorization.RequestParameters{
			QueryParameters: &map[string][]string{
				"param1": {"param-value-1"},
				"param2": {"param-value-2.1", "param-value-2.2"},
			},
		}

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:         ts.URL,
			Credentials:       credentials,
			RequestParameters: requestParameters,
		}, nil).Once()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		apiExtractorMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should proxy and add addidtional headers", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "custom-value-1", req.Header.Get("X-Custom1"))

			assert.Equal(t, 2, len(req.Header["X-Custom2"]))
			assert.Equal(t, "custom-value-2.1", req.Header.Get("X-Custom2"))
			assert.Equal(t, "custom-value-2.1", req.Header["X-Custom2"][0])
			assert.Equal(t, "custom-value-2.2", req.Header["X-Custom2"][1])
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil).
			Once()

		credentials := &authorization.Credentials{}
		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		requestParameters := &authorization.RequestParameters{
			Headers: &map[string][]string{
				"X-Custom1": {"custom-value-1"},
				"X-Custom2": {"custom-value-2.1", "custom-value-2.2"},
			},
		}

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:         ts.URL,
			Credentials:       credentials,
			RequestParameters: requestParameters,
		}, nil).Once()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should proxy and remove headers", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, "", req.Header.Get(httpconsts.HeaderXForwardedClientCert))
			assert.Equal(t, "", req.Header.Get(httpconsts.HeaderXForwardedFor))
			assert.Equal(t, "", req.Header.Get(httpconsts.HeaderXForwardedProto))
			assert.Equal(t, "", req.Header.Get(httpconsts.HeaderXForwardedHost))
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"
		req.Header.Set(httpconsts.HeaderXForwardedClientCert, "C=US;O=Example Organisation;CN=Test User 1")
		req.Header.Set(httpconsts.HeaderXForwardedFor, "client")
		req.Header.Set(httpconsts.HeaderXForwardedProto, "http")
		req.Header.Set(httpconsts.HeaderXForwardedHost, "demo.example.com")

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil).
			Once()

		credentials := &authorization.Credentials{}
		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:   ts.URL,
			Credentials: credentials,
		}, nil).Once()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		apiExtractorMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should proxy and use internal cache", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "orders/123", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil).
			Twice()

		credentials := &authorization.Credentials{}
		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", credentials).Return(authStrategyMock).Once()

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledTwice)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:   ts.URL,
			Credentials: credentials,
		}, nil).Once()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		// given
		nextReq, _ := http.NewRequest(http.MethodGet, "/orders/123", nil)
		nextReq.Host = "test-uuid-1.namespace.svc.cluster.local"
		rr = httptest.NewRecorder()

		//when
		handler.ServeHTTP(rr, nextReq)

		//then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
		apiExtractorMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
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

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil)

		credentialsMatcher := createOAuthCredentialsMatcher("clientId", "clientSecret", tsOAuth.URL+"/token")

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.MatchedBy(credentialsMatcher)).Return(authStrategyMock)

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl: ts.URL,
			Credentials: &authorization.Credentials{
				OAuth: &authorization.OAuth{
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
					URL:          tsOAuth.URL + "/token",
				},
			},
		}, nil)

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		apiExtractorMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should proxy BasicAuth auth calls", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(nil)

		credentialsMatcher := createBasicCredentialsMatcher("username", "password")

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.MatchedBy(credentialsMatcher)).Return(authStrategyMock)

		csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl: ts.URL,
			Credentials: &authorization.Credentials{
				BasicAuth: &authorization.BasicAuth{
					Username: "username",
					Password: "password",
				},
			},
		}, nil)

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		apiExtractorMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	t.Run("should fail with Bad Gateway error when failed to get OAuth token", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		req.Host = "test-uuid-1.namespace.svc.cluster.local"

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("TransportSetter")).
			Return(apperrors.UpstreamServerCallFailed("failed"))

		credentialsMatcher := createOAuthCredentialsMatcher("clientId", "clientSecret", "www.example.com/token")

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.MatchedBy(credentialsMatcher)).Return(authStrategyMock)
		csrfFactoryMock, csrfStrategyMock := neverCalledCSRFStrategy(authStrategyMock)

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl: ts.URL,
			Credentials: &authorization.Credentials{
				OAuth: &authorization.OAuth{
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
					URL:          "www.example.com/token",
				},
			},
		}, nil)

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadGateway, rr.Code)

		apiExtractorMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfFactoryMock.AssertExpectations(t)
		csrfStrategyMock.AssertExpectations(t)
	})

	testRetryOnAuthFailure := func(testServerConstructor func(check func(req *http.Request)) *httptest.Server, requestBody io.Reader, expectedStatusCode int, t *testing.T) {
		// given
		tsf := testServerConstructor(func(req *http.Request) {
			assertCookie(t, req, "user-cookie", "user-cookie-value")
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer tsf.Close()

		req, _ := http.NewRequest(http.MethodGet, "/orders/123", requestBody)
		req.Host = "test-uuid-1.namespace.svc.cluster.local"
		req.AddCookie(&http.Cookie{Name: "user-cookie", Value: "user-cookie-value"})

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:   tsf.URL,
			Credentials: &authorization.Credentials{},
		}, nil).Twice()

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.Anything, mock.AnythingOfType("TransportSetter")).
			Return(nil).Twice()
		authStrategyMock.On("Invalidate").Return().Once()

		csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
		csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).Return(nil).Twice()
		csrfTokenStrategyMock.On("Invalidate").Return().Once()

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.Anything).Return(authStrategyMock).Twice()

		csrfTokenStrategyFactoryMock := &csrfMock.TokenStrategyFactory{}
		csrfTokenStrategyFactoryMock.On("Create", authStrategyMock, "").Return(csrfTokenStrategyMock).Twice()

		handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfTokenStrategyFactoryMock, fakePathExtractor, createProxyConfig(proxyTimeout))
		rr := httptest.NewRecorder()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, expectedStatusCode, rr.Code)
		assert.Equal(t, "test", rr.Body.String())

		apiExtractorMock.AssertExpectations(t)
		authStrategyFactoryMock.AssertExpectations(t)
		authStrategyMock.AssertExpectations(t)
		csrfTokenStrategyFactoryMock.AssertExpectations(t)
		csrfTokenStrategyMock.AssertExpectations(t)
	}

	t.Run("should invalidate proxy and retry when 401 occurred", func(t *testing.T) {
		testRetryOnAuthFailure(func(check func(req *http.Request)) *httptest.Server {
			return NewTestServerForRetryTest(http.StatusUnauthorized, check)
		}, nil, http.StatusOK, t)
	})

	t.Run("should invalidate proxy and retry when 403 occurred due to CRSF Token validation", func(t *testing.T) {
		testRetryOnAuthFailure(func(check func(req *http.Request)) *httptest.Server {
			return NewTestServerForRetryTest(http.StatusForbidden, check)
		}, nil, http.StatusOK, t)
	})

	t.Run("should return 403 status when the call and the retry with body returned 403", func(t *testing.T) {
		requestBody := bytes.NewBufferString("some body")
		testRetryOnAuthFailure(func(check func(req *http.Request)) *httptest.Server {
			return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.ParseForm()
				check(r)
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("test"))
			}))

		}, requestBody, http.StatusForbidden, t)
	})
}

func assertCookie(t *testing.T, r *http.Request, name, value string) {
	cookie, err := r.Cookie(name)
	require.NoError(t, err)

	assert.Equal(t, value, cookie.Value)
}

func NewTestServer(check func(req *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
}

func newProxyForTest(
	apiExtractor APIExtractor,
	authorizationStrategyFactory authorization.StrategyFactory,
	csrfTokenStrategyFactory csrf.TokenStrategyFactory,
	pathExtractorFunc pathExtractorFunc,
	proxyConfig Config) http.Handler {

	return &proxy{
		cache:                        NewCache(proxyConfig.ProxyCacheTTL),
		skipVerify:                   proxyConfig.SkipVerify,
		proxyTimeout:                 proxyConfig.ProxyTimeout,
		authorizationStrategyFactory: authorizationStrategyFactory,
		csrfTokenStrategyFactory:     csrfTokenStrategyFactory,
		extractPathFunc:              pathExtractorFunc,
		apiExtractor:                 apiExtractor,
	}
}

func NewTestServerForRetryTest(status int, check func(req *http.Request)) *httptest.Server {
	willFail := true

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		check(r)
		if willFail {
			w.WriteHeader(status)
			willFail = false
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte("test"))
	}))
}

func createProxyConfig(proxyTimeout int) Config {
	return Config{
		SkipVerify:    true,
		ProxyTimeout:  proxyTimeout,
		Application:   "test",
		ProxyCacheTTL: proxyTimeout,
	}
}

func createOAuthCredentialsMatcher(clientID, clientSecret, url string) func(*authorization.Credentials) bool {
	return func(c *authorization.Credentials) bool {
		return c.OAuth != nil && c.OAuth.ClientID == clientID &&
			c.OAuth.ClientSecret == clientSecret &&
			c.OAuth.URL == url
	}
}

func createBasicCredentialsMatcher(username, password string) func(*authorization.Credentials) bool {
	return func(c *authorization.Credentials) bool {
		return c.BasicAuth != nil && c.BasicAuth.Username == username &&
			c.BasicAuth.Password == password
	}
}

func mockCSRFStrategy(authorizationStrategy authorization.Strategy, ef ensureCalledFunc) (*csrfMock.TokenStrategyFactory, *csrfMock.TokenStrategy) {

	csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
	strategyCall := csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).
		Return(nil)
	ef(strategyCall)

	csrfTokenStrategyFactoryMock := &csrfMock.TokenStrategyFactory{}
	csrfTokenStrategyFactoryMock.On("Create", authorizationStrategy, "").Return(csrfTokenStrategyMock).Once()

	return csrfTokenStrategyFactoryMock, csrfTokenStrategyMock
}

func neverCalledCSRFStrategy(authorizationStrategy authorization.Strategy) (*csrfMock.TokenStrategyFactory, *csrfMock.TokenStrategy) {
	csrfTokenStrategyMock := &csrfMock.TokenStrategy{}

	csrfTokenStrategyFactoryMock := &csrfMock.TokenStrategyFactory{}
	csrfTokenStrategyFactoryMock.On("Create", authorizationStrategy, "").Return(csrfTokenStrategyMock).Once()

	return csrfTokenStrategyFactoryMock, csrfTokenStrategyMock
}

type ensureCalledFunc func(mockCall *mock.Call)

func calledTwice(mockCall *mock.Call) {
	mockCall.Twice()
}

func calledOnce(mockCall *mock.Call) {
	mockCall.Once()
}

func readErrorResponse(t *testing.T, body io.Reader) httperrors.ErrorResponse {
	responseBody, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var errorResponse httperrors.ErrorResponse
	err = json.Unmarshal(responseBody, &errorResponse)
	require.NoError(t, err)

	return errorResponse
}
