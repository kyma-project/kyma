package client

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/testconsts"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	timeoutDuration = 5
	testURL         = "test.io/token"

	cachedTestToken            = "someToken"
	cachedTestCookieName       = "someCookie"
	endpointTestToken          = "someEndpointToken"
	endpointResponseCookieName = "someOtherCookie"

	testUsername          = "someUser"
	testPassword          = "somePassword"
	expectedAuthHeaderVal = "Basic c29tZVVzZXI6c29tZVBhc3N3b3Jk"
)

var (
	certificate = []byte(testconsts.Certificate)
	privateKey  = []byte(testconsts.PrivateKey)
)

func TestClient_GetTokenEndpointResponse(t *testing.T) {

	sf := authorization.NewStrategyFactory(authorization.FactoryConfiguration{OAuthClientTimeout: timeoutDuration})

	strategy := sf.Create(&authorization.Credentials{BasicAuth: &authorization.BasicAuth{
		Username: testUsername,
		Password: testPassword,
	}})

	t.Run("Should fetch the token from cache if it is present", func(t *testing.T) {

		// given
		r := &csrf.Response{
			CSRFToken: cachedTestToken,
			Cookies:   []*http.Cookie{{Name: cachedTestCookieName}},
		}

		fakeCache := NewTokenCache()
		fakeCache.Add(testURL, r)

		c := client{timeoutDuration, fakeCache, nil}

		// when
		response, appError := c.GetTokenEndpointResponse(testURL, nil)

		// then
		require.Nil(t, appError)
		require.NotNil(t, response)

		assert.Equal(t, cachedTestToken, response.CSRFToken)
		assert.Equal(t, cachedTestCookieName, response.Cookies[0].Name)
	})

	t.Run("Should fetch the token from endpoint and add it to cache if it is not there", func(t *testing.T) {

		// given
		fakeCache := NewTokenCache()

		c := client{timeoutDuration, fakeCache, &http.Client{}}

		srv := startTestServer(t)
		mockURL := srv.URL

		// when
		response, appError := c.GetTokenEndpointResponse(mockURL, strategy)
		item, found := fakeCache.Get(mockURL)

		// then
		require.Nil(t, appError)
		require.NotNil(t, response)

		assert.Equal(t, endpointTestToken, response.CSRFToken)
		assert.Equal(t, endpointResponseCookieName, response.Cookies[0].Name)

		assert.True(t, found)
		require.NotNil(t, item)
		assert.Equal(t, endpointTestToken, item.CSRFToken)
		assert.Equal(t, endpointResponseCookieName, item.Cookies[0].Name)
	})

	t.Run("Should return error if the token requested token is not in the cache and can't be retrieved", func(t *testing.T) {

		// given
		fakeCache := NewTokenCache()

		c := client{timeoutDuration, fakeCache, &http.Client{}}

		srv := startFailingTestServer(t)
		mockURL := srv.URL

		// when
		response, appError := c.GetTokenEndpointResponse(mockURL, strategy)
		item, found := fakeCache.Get(mockURL)

		// then
		require.NotNil(t, appError)
		require.Nil(t, response)

		require.Nil(t, item)
		assert.False(t, found)
	})
}

func TestAddAuthorization(t *testing.T) {

	sf := authorization.NewStrategyFactory(authorization.FactoryConfiguration{OAuthClientTimeout: timeoutDuration})

	t.Run("Should update request with authorization headers in case of basicAuth strategy", func(t *testing.T) {

		// given
		strategy := sf.Create(&authorization.Credentials{BasicAuth: &authorization.BasicAuth{
			Username: testUsername,
			Password: testPassword,
		}})

		client := &http.Client{}
		request := getNewEmptyRequest()

		// when
		addAuthorization(request, client, strategy)

		// then
		assert.Len(t, request.Header, 1)
		assert.NotEmpty(t, request.Header.Get(httpconsts.HeaderAuthorization))
		assert.Equal(t, expectedAuthHeaderVal, request.Header.Get(httpconsts.HeaderAuthorization))
	})

	t.Run("Should update httpClient with transport in case of certificateGen strategy", func(t *testing.T) {

		// given
		strategy := sf.Create(&authorization.Credentials{CertificateGen: &authorization.CertificateGen{
			CommonName:  "",
			PrivateKey:  privateKey,
			Certificate: certificate,
		}})

		client := &http.Client{}
		request := getNewEmptyRequest()

		// when
		addAuthorization(request, client, strategy)

		// then
		assert.NotNil(t, client.Transport)
	})
}

func TestSetCSRFSpecificHeaders(t *testing.T) {

	t.Run("Should add CSRF specific headers to the request", func(t *testing.T) {

		// given
		r := getNewEmptyRequest()

		// when
		setCSRFSpecificHeaders(r)

		// then
		assert.Len(t, r.Header, 3)
		assert.NotEmpty(t, r.Header.Get(httpconsts.HeaderCSRFToken))
		assert.NotEmpty(t, r.Header.Get(httpconsts.HeaderAccept))
		assert.NotEmpty(t, r.Header.Get(httpconsts.HeaderCacheControl))
	})
}

func getNewEmptyRequest() *http.Request {
	return &http.Request{
		Header: make(map[string][]string),
	}
}

func startTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkRequest(t, r)
		w.Header().Add("x-csrf-token", endpointTestToken)
		http.SetCookie(w, &http.Cookie{Name: endpointResponseCookieName})
		w.WriteHeader(http.StatusOK)
	}))
}

func startFailingTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkRequest(t, r)
		w.WriteHeader(http.StatusNotFound)
	}))
}

func checkRequest(t *testing.T, r *http.Request) {
	authHeader := r.Header.Get(httpconsts.HeaderAuthorization)
	encodedCredentials := strings.TrimPrefix(string(authHeader), "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	require.NoError(t, err)
	credentials := strings.Split(string(decoded), ":")

	assert.Equal(t, testUsername, credentials[0])
	assert.Equal(t, testPassword, credentials[1])

	assert.NotEmpty(t, r.Header.Get(httpconsts.HeaderCSRFToken))
	assert.NotEmpty(t, r.Header.Get(httpconsts.HeaderAccept))
	assert.NotEmpty(t, r.Header.Get(httpconsts.HeaderCacheControl))
}
