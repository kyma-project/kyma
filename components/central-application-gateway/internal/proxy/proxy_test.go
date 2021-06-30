package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"

	csrfMock "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/mocks"
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

func TestProxyRequest(t *testing.T) {
	forbiddenHeaders := []string{
		httpconsts.HeaderXForwardedClientCert,
		httpconsts.HeaderXForwardedFor,
		httpconsts.HeaderXForwardedProto,
		httpconsts.HeaderXForwardedHost,
	}
	type apiExtractor struct {
		targetPath        string
		requestParameters *authorization.RequestParameters
		credentials       authorization.Credentials
	}
	type request struct {
		url    string
		header http.Header
	}
	type expectedProxyRequest struct {
		targetUrl string
		header    http.Header
	}
	type testcase struct {
		name                 string
		request              request
		apiExtractor         apiExtractor
		expectedProxyRequest expectedProxyRequest
	}
	tests := []testcase{
		{
			name: "Should proxy without escaping the URL path characters when target URL does not contain path",
			request: request{
				url: "/somepath/Xyz('123')",
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/somepath/Xyz('123')",
			},
		},
		{
			name: "should proxy without escaping the URL path characters when target URL contains path",
			request: request{
				url: "/Xyz('123')",
			},
			apiExtractor: apiExtractor{
				targetPath: "/somepath",
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/somepath/Xyz('123')",
			},
		},
		{
			name: "should proxy without escaping the URL path characters when target URL contains full path",
			request: request{
				url: "?$search=XXX",
			},
			apiExtractor: apiExtractor{
				targetPath: "/somepath/Xyz('123')",
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/somepath/Xyz('123')?$search=XXX",
			},
		},
		{
			name: "should proxy and add additional query parameters",
			request: request{
				url: "/orders/123",
			},
			apiExtractor: apiExtractor{
				requestParameters: &authorization.RequestParameters{
					QueryParameters: &map[string][]string{
						"param1": {"param-value-1"},
						"param2": {"param-value-2.1", "param-value-2.2"},
					},
				},
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/orders/123?param1=param-value-1&param2=param-value-2.1&param2=param-value-2.2",
			},
		},
		{
			name: "should proxy and add additional headers",
			request: request{
				url: "/orders/123",
				header: map[string][]string{
					"X-Request1": {"request-value-1"},
					"X-Request2": {"request-value-2.1", "request-value-2.2"},
				},
			},
			apiExtractor: apiExtractor{
				requestParameters: &authorization.RequestParameters{
					Headers: &map[string][]string{
						"X-Custom1": {"custom-value-1"},
						"X-Custom2": {"custom-value-2.1", "custom-value-2.2"},
					},
				},
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/orders/123",
				header: map[string][]string{
					"X-Custom1":  {"custom-value-1"},
					"X-Custom2":  {"custom-value-2.1", "custom-value-2.2"},
					"X-Request1": {"request-value-1"},
					"X-Request2": {"request-value-2.1", "request-value-2.2"},
				},
			},
		},
		{
			name: "should proxy and remove headers",
			request: request{
				url: "/orders/123",
				header: map[string][]string{
					httpconsts.HeaderXForwardedClientCert: {"C=US;O=Example Organisation;CN=Test User 1"},
					httpconsts.HeaderXForwardedFor:        {"client"},
					httpconsts.HeaderXForwardedProto:      {"http"},
					httpconsts.HeaderXForwardedHost:       {"demo.example.com"},
				},
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/orders/123",
			},
		},
		{
			name: "should proxy BasicAuth auth calls",
			request: request{
				url: "/orders/123",
			},
			apiExtractor: apiExtractor{
				credentials: authorization.Credentials{
					BasicAuth: &authorization.BasicAuth{
						Username: "username",
						Password: "password",
					},
				},
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/orders/123",
				// authorization header is not set by the mock
			},
		},
		{
			name: "should proxy OAuth calls",
			request: request{
				url: "/orders/123",
			},
			apiExtractor: apiExtractor{
				credentials: authorization.Credentials{
					OAuth: &authorization.OAuth{
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
						URL:          "www.example.com/token",
					},
				},
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/orders/123",
				// authorization header is not set by the mock
			},
		},
		{
			name: "should fail with Bad Gateway error when failed to get OAuth token",
			request: request{
				url: "/orders/123",
			},
			apiExtractor: apiExtractor{
				credentials: authorization.Credentials{
					OAuth: &authorization.OAuth{
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
						URL:          "www.example.com/token",
					},
				},
			},
			expectedProxyRequest: expectedProxyRequest{
				targetUrl: "/orders/123",
				// authorization header is not set by the mock
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			ts := NewTestServer(func(req *http.Request) {
				expectedUrl, err := url.Parse(tc.expectedProxyRequest.targetUrl)
				assert.Nil(t, err)
				// compare maps objects, rather than strings built from unordered maps
				assert.Equal(t, expectedUrl.Query(), req.URL.Query())
				assert.Equal(t, expectedUrl.RequestURI(), req.URL.RequestURI())

				for name, values := range tc.expectedProxyRequest.header {
					assert.Equal(t, values, req.Header.Values(name))
				}
				for _, name := range forbiddenHeaders {
					assert.Equal(t, "", req.Header.Get(name))
				}
			})
			defer ts.Close()

			req, err := http.NewRequest(http.MethodGet, tc.request.url, nil)
			require.NoError(t, err)
			for name, values := range tc.request.header {
				req.Header[name] = values
			}
			authStrategyMock := &authMock.Strategy{}
			authStrategyMock.
				On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("SetClientCertificateFunc")).
				Return(nil).
				Once()

			credentialsMatcher := createCredentialsMatcher(&tc.apiExtractor.credentials)
			authStrategyFactoryMock := &authMock.StrategyFactory{}
			authStrategyFactoryMock.On("Create", mock.MatchedBy(credentialsMatcher)).Return(authStrategyMock).Once()

			csrfFactoryMock, csrfStrategyMock := mockCSRFStrategy(authStrategyMock, calledOnce)

			apiExtractorMock := &proxyMocks.APIExtractor{}
			apiExtractorMock.On("Get", metadatamodel.APIIdentifier{
				Application: "app",
				Service:     "service",
				Entry:       "entry",
			}).Return(&metadatamodel.API{
				TargetUrl:         ts.URL + tc.apiExtractor.targetPath,
				Credentials:       &tc.apiExtractor.credentials,
				RequestParameters: tc.apiExtractor.requestParameters,
			}, nil).Once()

			handler := newProxyForTest(apiExtractorMock, authStrategyFactoryMock, csrfFactoryMock, func(path string) (metadatamodel.APIIdentifier, string, apperrors.AppError) {
				return metadatamodel.APIIdentifier{
					Application: "app",
					Service:     "service",
					Entry:       "entry",
				}, path, nil
			}, createProxyConfig(10))
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
	}
}

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

	t.Run("should fail with Bad Gateway error when failed to get OAuth token", func(t *testing.T) {
		// given
		ts := NewTestServer(func(req *http.Request) {
			assert.Equal(t, req.Method, http.MethodGet)
			assert.Equal(t, req.RequestURI, "/orders/123")
		})
		defer ts.Close()

		req, err := http.NewRequest(http.MethodGet, "/orders/123", nil)
		require.NoError(t, err)

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.AnythingOfType("*http.Request"), mock.AnythingOfType("SetClientCertificateFunc")).
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
		req.AddCookie(&http.Cookie{Name: "user-cookie", Value: "user-cookie-value"})

		apiExtractorMock := &proxyMocks.APIExtractor{}
		apiExtractorMock.On("Get", apiIdentifier).Return(&metadatamodel.API{
			TargetUrl:   tsf.URL,
			Credentials: &authorization.Credentials{},
		}, nil)

		authStrategyMock := &authMock.Strategy{}
		authStrategyMock.
			On("AddAuthorization", mock.Anything, mock.AnythingOfType("SetClientCertificateFunc")).
			Return(nil).Twice()
		authStrategyMock.On("Invalidate").Return().Once()

		csrfTokenStrategyMock := &csrfMock.TokenStrategy{}
		csrfTokenStrategyMock.On("AddCSRFToken", mock.AnythingOfType("*http.Request")).Return(nil).Twice()
		csrfTokenStrategyMock.On("Invalidate").Return().Once()

		authStrategyFactoryMock := &authMock.StrategyFactory{}
		authStrategyFactoryMock.On("Create", mock.Anything).Return(authStrategyMock)

		csrfTokenStrategyFactoryMock := &csrfMock.TokenStrategyFactory{}
		csrfTokenStrategyFactoryMock.On("Create", authStrategyMock, "").Return(csrfTokenStrategyMock)

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

type CredentialsMatcherFunc func(*authorization.Credentials) bool

func createCredentialsMatcher(creds *authorization.Credentials) CredentialsMatcherFunc {
	if creds.BasicAuth != nil {
		return createBasicCredentialsMatcher(creds.BasicAuth.Username, creds.BasicAuth.Password)
	}
	if creds.OAuth != nil {
		return createOAuthCredentialsMatcher(creds.OAuth.ClientID, creds.OAuth.ClientSecret, creds.OAuth.URL)
	}
	return createEmptyCredentialsMatcher()
}

func createEmptyCredentialsMatcher() CredentialsMatcherFunc {
	return func(c *authorization.Credentials) bool {
		return c != nil && *c == authorization.Credentials{}
	}
}

func createOAuthCredentialsMatcher(clientID, clientSecret, url string) CredentialsMatcherFunc {
	return func(c *authorization.Credentials) bool {
		return c.OAuth != nil && c.OAuth.ClientID == clientID &&
			c.OAuth.ClientSecret == clientSecret &&
			c.OAuth.URL == url
	}
}

func createBasicCredentialsMatcher(username, password string) CredentialsMatcherFunc {
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

func calledOnce(mockCall *mock.Call) {
	mockCall.Once()
}
