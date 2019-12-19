package oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/oauth/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOauthClient_GetToken(t *testing.T) {
	t.Run("should get token from cache if present", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID").Return("123456789", true)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", "", nil, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fetch token from EC when token if not present in cache", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checkAccessTokenRequest(t, r)

			response := oauthResponse{AccessToken: "123456789", TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID").Return("", false)
		tokenCache.On("Add", "testID", "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fetch token using also additional headers and query parameters", func(t *testing.T) {
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

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID").Return("", false)
		tokenCache.On("Add", "testID", "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, &headers, &queryParameters)

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

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID").Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil)

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

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID").Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", ts.URL, nil, nil)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if OAuth address is incorrect", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "testID").Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("testID", "testSecret", "http://some_no_existent_address.com/token", nil, nil)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})
}

func TestOauthClient_InvalidateAndRetry(t *testing.T) {
	t.Run("should fetch token from EC", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checkAccessTokenRequest(t, r)

			response := oauthResponse{AccessToken: "123456789", TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "testID")
		tokenCache.On("Add", "testID", "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("testID", "testSecret", ts.URL, nil, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, "123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail when unable to get token", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			checkAccessTokenRequest(t, r)

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "testID")

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("testID", "testSecret", ts.URL, nil, nil)

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

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "testID")

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("testID", "testSecret", ts.URL, nil, nil)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if OAuth address is incorrect", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "testID")

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("testID", "testSecret", "http://some_no_existent_address.com/token", nil, nil)

		// then
		require.Error(t, err)
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
