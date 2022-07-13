package test_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

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

type OAuthHandler struct {
	clientID     string
	clientSecret string
	mutex        sync.RWMutex
	tokens       map[string]bool
}

func NewOAuth(clientID, clientSecret string) OAuthHandler {
	return OAuthHandler{
		clientID:     clientID,
		clientSecret: clientSecret,
		mutex:        sync.RWMutex{},
		tokens:       make(map[string]bool),
	}
}

func (oh *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse form: %v", err))
		return
	}

	clientID := r.FormValue(clientIDKey)
	clientSecret := r.FormValue(clientSecretKey)
	grantType := r.FormValue(grantTypeKey)

	if !oh.verifyClient(clientID, clientSecret) || grantType != "client_credentials" {
		handleError(w, http.StatusForbidden, "Client verification failed")
		return
	}

	token := uuid.New().String()

	oh.mutex.Lock()
	oh.tokens[token] = true
	oh.mutex.Unlock()

	response := OauthResponse{AccessToken: token, TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, http.StatusInternalServerError, "Failed to encode token response")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (oh *OAuthHandler) BadToken(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse form: %v", err))
		return
	}

	clientID := r.FormValue(clientIDKey)
	clientSecret := r.FormValue(clientSecretKey)
	grantType := r.FormValue(grantTypeKey)

	if !oh.verifyClient(clientID, clientSecret) || grantType != "client_credentials" {
		handleError(w, http.StatusForbidden, "Client verification failed")
		return
	}

	token := uuid.New().String()

	response := OauthResponse{AccessToken: token, TokenType: "bearer", ExpiresIn: 3600, Scope: "basic"}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, http.StatusInternalServerError, "Failed to encode token response")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (oh *OAuthHandler) Middleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handleError(w, http.StatusForbidden, "Authorization header missing")
				return
			}

			splitToken := strings.Split(authHeader, "Bearer")
			if len(splitToken) != 2 {
				handleError(w, http.StatusForbidden, "Bearer token missing")
				return
			}

			token := strings.TrimSpace(splitToken[1])

			oh.mutex.RLock()
			_, found := oh.tokens[token]
			oh.mutex.RUnlock()

			if !found {
				handleError(w, http.StatusForbidden, "Invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (oh *OAuthHandler) verifyClient(id, secret string) bool {
	return id == oh.clientID && secret == oh.clientSecret
}
