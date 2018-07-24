package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToken(t *testing.T) {

	t.Run("should fetch token from EC", func(t *testing.T) {
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

		oauthClient := NewOauthClient(10)
		token, err := oauthClient.GetToken("test", "test", ts.URL)

		require.NoError(t, err)
		assert.Equal(t, "Bearer 123456789", token)
	})

	t.Run("should fail when unable to get token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		oauthClient := NewOauthClient(10)
		token, err := oauthClient.GetToken("test", "test", ts.URL)

		require.Error(t, err)
		assert.Equal(t, "", token)
	})

	t.Run("should fail if payload is empty", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r.ParseForm()

			assert.Equal(t, "test", r.PostForm.Get("client_id"))
			assert.Equal(t, "test", r.PostForm.Get("client_secret"))
			assert.Equal(t, "client_credentials", r.PostForm.Get("grant_type"))

			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		oauthClient := NewOauthClient(10)
		token, err := oauthClient.GetToken("test", "test", ts.URL)

		require.Error(t, err)
		assert.Equal(t, "", token)
	})

	t.Run("should fail if OAuth address is incorrect", func(t *testing.T) {

		oauthClient := NewOauthClient(10)
		token, err := oauthClient.GetToken("test", "test", "http://some_no_existent_address.com/token")

		require.Error(t, err)
		assert.Equal(t, "", token)
	})

}
