package test_api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

func AddTokensHandler(router *mux.Router, oauthTokesCache map[string]bool, csrfTokensCache map[string]bool, credentials OAuthCredentials, mutex *sync.RWMutex) {
	tokensHandler := router

	tokensHandler.HandleFunc("/v1/server/oauth/token", NewOAuthServerHandler(oauthTokesCache, credentials, mutex)).Methods("POST")
}

type OAuthCredentials struct {
	ClientID     string
	ClientSecret string
}

const (
	clientIDKey     = "client_id"
	clientSecretKey = "client_secret"
	grantTypeKey    = "grant_type"
)

type OauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func NewOAuthServerHandler(oauthTokesCache map[string]bool, credentials OAuthCredentials, mutex *sync.RWMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			handleError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse form: %v", err))
			return
		}

		clientID := r.FormValue(clientIDKey)
		clientSecret := r.FormValue(clientSecretKey)
		grantType := r.FormValue(grantTypeKey)

		if clientID != credentials.ClientID || clientSecret != credentials.ClientSecret || grantType != "client_credentials" {
			handleError(w, http.StatusForbidden, "Invalid token")
			return
		}

		token := uuid.New().String()

		// We could skip locking as we don't expect multiple clients to access this service.
		mutex.Lock()
		oauthTokesCache[token] = true
		mutex.Unlock()

		response := OauthResponse{AccessToken: token, TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			handleError(w, http.StatusInternalServerError, "Failed to encode token response")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
