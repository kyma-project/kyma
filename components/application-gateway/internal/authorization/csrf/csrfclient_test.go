package csrf

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization/testconsts"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	timeoutDuration = 5
	testURL         = "someURL"

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

	t.Run("Should fetch the token from cache if it is present", func(t *testing.T) {

		// given
		r := &Response{
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

		c := client{timeoutDuration, fakeCache, testRequestToken}

		// when
		response, appError := c.GetTokenEndpointResponse(testURL, nil)
		item, found := fakeCache.Get(testURL)

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

		c := client{timeoutDuration, fakeCache, failingTestRequestToken}

		// when
		response, appError := c.GetTokenEndpointResponse(testURL, nil)
		item, found := fakeCache.Get(testURL)

		// then
		require.NotNil(t, appError)
		require.Nil(t, response)

		require.Nil(t, item)
		assert.False(t, found)
	})
}

func testRequestToken(_ string, _ authorization.Strategy, _ int) (*Response, apperrors.AppError) {
	return &Response{
		CSRFToken: endpointTestToken,
		Cookies:   []*http.Cookie{{Name: endpointResponseCookieName}},
	}, nil
}

func failingTestRequestToken(_ string, _ authorization.Strategy, _ int) (*Response, apperrors.AppError) {
	return nil, apperrors.NotFound("")
}

func TestAddAuthorization(t *testing.T) {

	sf := authorization.NewStrategyFactory(authorization.FactoryConfiguration{timeoutDuration})

	t.Run("Should update request with authorization headers in case of basicAuth strategy", func(t *testing.T) {

		// given
		strategy := sf.Create(&model.Credentials{BasicAuth: &model.BasicAuth{
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

	t.Run("Should update client with transport in case of certificateGen strategy", func(t *testing.T) {

		// given
		strategy := sf.Create(&model.Credentials{CertificateGen: &model.CertificateGen{
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
