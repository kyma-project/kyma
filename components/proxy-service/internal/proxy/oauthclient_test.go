package proxy

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/proxy-service/internal/proxy/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOauthClient_GetToken(t *testing.T) {
	t.Run("should get token from cache if present", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "test").Return("123456789", true)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("test", "test", "")

		// then
		require.NoError(t, err)
		assert.Equal(t, "Bearer 123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fetch token from EC when token if not present in cache", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			response := oauthResponse{AccessToken: "123456789", TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "test").Return("", false)
		tokenCache.On("Add", "test", "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("test", "test", ts.URL)

		// then
		require.NoError(t, err)
		assert.Equal(t, "Bearer 123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail when unable to get token", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "test").Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("test", "test", ts.URL)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if payload is empty", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "test").Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("test", "test", ts.URL)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if OAuth address is incorrect", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Get", "test").Return("", false)

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.GetToken("test", "test", "http://some_no_existent_address.com/token")

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

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			response := oauthResponse{AccessToken: "123456789", TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "test")
		tokenCache.On("Add", "test", "123456789", 3600).Return()

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("test", "test", ts.URL)

		// then
		require.NoError(t, err)
		assert.Equal(t, "Bearer 123456789", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail when unable to get token", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "test")

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("test", "test", ts.URL)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if payload is empty", func(t *testing.T) {
		// given
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "test")

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("test", "test", ts.URL)

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})

	t.Run("should fail if OAuth address is incorrect", func(t *testing.T) {
		// given
		tokenCache := mocks.TokenCache{}
		tokenCache.On("Remove", "test")

		oauthClient := NewOauthClient(10, &tokenCache)

		// when
		token, err := oauthClient.InvalidateAndRetry("test", "test", "http://some_no_existent_address.com/token")

		// then
		require.Error(t, err)
		assert.Equal(t, "", token)
		tokenCache.AssertExpectations(t)
	})
}
