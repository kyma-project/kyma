package oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOauthClient_GetToken(t *testing.T) {
	t.Run("should get token from cache if present", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testIDtestSecret").Return("123456789", true)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", "", nil, nil, false)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fetch token from server when token if not present in cache", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checkAccessTokenRequest(t, r)

			response := oauthResponse{AccessToken: "123456789", TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		tokenKey := "testID" + "testSecret" + ts.URL

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", tokenKey).Return("", false)
		tokenCache.On("Add", tokenKey, "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil, false)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fetch token from insecure server when token if not present in cache", func(t *testing.T) {
		// given
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checkAccessTokenRequest(t, r)

			response := oauthResponse{AccessToken: "123456789", TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))

		ts.StartTLS()
		defer ts.Close()

		tokenKey := "testID" + "testSecret" + ts.URL

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", tokenKey).Return("", false)
		tokenCache.On("Add", tokenKey, "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fetch token using additional headers and query parameters", func(t *testing.T) {
		// given
		headers := map[string][]string{
			"headerKey": {"headerValue"},
		}
		queryParameters := map[string][]string{
			"queryParameterKey": {"queryParameterValue"},
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checkAccessTokenRequest(t, r)
			checkAccessTokenRequestAdditionalRequestParameters(t, r)

			response := oauthResponse{AccessToken: "123456789", TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		tokenKey := "testID" + "testSecret" + ts.URL

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", tokenKey).Return("", false)
		tokenCache.On("Add", tokenKey, "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, &headers, &queryParameters, false)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail when unable to get token", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		tokenKey := "testID" + "testSecret" + ts.URL

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", tokenKey).Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil, false)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if payload is empty", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checkAccessTokenRequest(t, r)

			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		tokenKey := "testID" + "testSecret" + ts.URL

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", tokenKey).Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil, false)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if OAuth address is incorrect", func(t *testing.T) {
		// given
		tokenKey := "testID" + "testSecret" + "http://some_no_existent_address.com/token"

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", tokenKey).Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", "http://some_no_existent_address.com/token", nil, nil, false)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail when calling server protected with self-signed certificate", func(t *testing.T) {
		// given
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ts.StartTLS()
		defer ts.Close()

		tokenKey := "testID" + "testSecret" + ts.URL

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", tokenKey).Return("", false)
		//tokenCache.On("Add", mock.Anything, mock.Anything, mock.Anything).Times(0)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		_, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil, false)

		// then
		require.Error(t, err)
		tokenCache.AssertExpectations(t)
	})
}

func TestOauthClient_GetTokenMTLS(t *testing.T) {
	var certSHA = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	var keySHA = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"

	t.Run("should get token from cache if present", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID-"+certSHA+"-"+keySHA+"-testURL").Return("123456789", true)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetTokenMTLS("testID", "testURL", []byte("test"), []byte("test"), nil, nil, false)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if Certificate and Private Key is not valid", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID-"+certSHA+"-"+keySHA+"-testURL").Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetTokenMTLS("testID", "testURL", []byte("test"), []byte("test"), nil, nil, false)

		// then
		assert.Error(t, err, apperrors.Internal("Failed to prepare certificate, %s", err.Error()))
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})
}

func checkAccessTokenRequest(t *testing.T, r *http.Request) {
	err := r.ParseForm()
	require.NoError(t, err)

	assert.Equal(t, "testID", r.PostForm.Get("client_id"))
	assert.Equal(t, "testSecret", r.PostForm.Get("client_secret"))
	assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

	authHeader := r.Header.Get(httpconsts.HeaderAuthorization)
	encodedCredentials := strings.TrimPrefix(string(authHeader), "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	require.NoError(t, err)
	credentials := strings.Split(string(decoded), ":")
	assert.Equal(t, "testID", credentials[0])
	assert.Equal(t, "testSecret", credentials[1])
}

func checkAccessTokenRequestAdditionalRequestParameters(t *testing.T, r *http.Request) {
	assert.Equal(t, []string{"queryParameterValue"}, r.URL.Query()["queryParameterKey"])
	assert.Equal(t, "headerValue", r.Header.Get("headerKey"))
}
