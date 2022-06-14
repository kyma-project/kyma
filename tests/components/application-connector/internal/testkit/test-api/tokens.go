package test_api

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

func AddTokensHandler(router *mux.Router, oauthTokesCache map[string]bool, csrfTokensCache map[string]bool, credentials OAuthCredentials, mutex *sync.RWMutex) {
	tokensHandler := router

	tokensHandler.HandleFunc("/v1/server/oauth/token", NewOAuthServerHandler(oauthTokesCache, credentials, mutex)).Methods("GET")
}

type OAuthCredentials struct {
	ClientID     string
	ClientSecret string
}

const (
	clientID     = "client_id"
	clientSecret = "client_secret"
	grantType    = "grant_type"
)

func NewOAuthServerHandler(oauthTokesCache map[string]bool, credentials OAuthCredentials, mutex *sync.RWMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var clientID, clientSecret, grantType string

		for key, value := range r.Form {
			if key == clientID {
				clientID = value[0]
			}

			if key == clientSecret {
				clientSecret = value[0]
			}

			if key == grantType {
				grantType = value[0]
			}
		}

		if clientID != credentials.ClientID || clientSecret != credentials.ClientSecret || grantType != "client_credentials" {
			handleError(w, 403, "Invalid token")
			return
		}

		token := uuid.New().String()

		// We could skip locking as we don't expect multiple clients to access this service.
		mutex.Lock()
		defer mutex.Unlock()
		oauthTokesCache[token] = true
	}
}
