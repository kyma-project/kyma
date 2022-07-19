package test_api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

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
	tokenLifetime   = "token_lifetime"
)

const (
	CtxOAuthToken ContextKey = iota
)

type OauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in,omitempty"`
}

type OAuthToken struct {
	exp time.Time
}

func (token OAuthToken) Valid() bool {
	return token.exp.After(time.Now())
}

type OAuthHandler struct {
	clientID     string
	clientSecret string
	mutex        sync.RWMutex
	tokens       map[string]OAuthToken
}

func NewOAuth(clientID, clientSecret string) OAuthHandler {
	return OAuthHandler{
		clientID:     clientID,
		clientSecret: clientSecret,
		mutex:        sync.RWMutex{},
		tokens:       make(map[string]OAuthToken),
	}
}

func (oh *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	if ok, status, message := oh.isRequestValid(r); !ok {
		handleError(w, status, message)
		return
	}

	token := uuid.New().String()
	ttl := 5 * time.Minute

	if ttlStr := r.URL.Query().Get(tokenLifetime); ttlStr != "" {
		if ttlOrErr, err := time.ParseDuration(ttlStr); err == nil { // TODO: Nesting ugly 🤮
			ttl = ttlOrErr
		}
	}

	oh.mutex.Lock()
	oh.tokens[token] = OAuthToken{exp: time.Now().Add(ttl)}
	oh.mutex.Unlock()

	response := OauthResponse{AccessToken: token, TokenType: "bearer", ExpiresIn: int64(ttl.Seconds())}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, http.StatusInternalServerError, "Failed to encode token response")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (oh *OAuthHandler) BadToken(w http.ResponseWriter, r *http.Request) {
	if ok, status, message := oh.isRequestValid(r); !ok {
		handleError(w, status, message)
		return
	}

	token := uuid.New().String()

	response := OauthResponse{AccessToken: token, TokenType: "bearer"}

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
				handleError(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			splitToken := strings.Split(authHeader, "Bearer")
			if len(splitToken) != 2 {
				handleError(w, http.StatusUnauthorized, "Bearer token missing")
				return
			}

			token := strings.TrimSpace(splitToken[1])

			oh.mutex.RLock()
			data, found := oh.tokens[token]
			oh.mutex.RUnlock()

			if !found || !data.Valid() {
				handleError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), CtxOAuthToken, token)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (oh *OAuthHandler) Deauth(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(CtxOAuthToken).(string)
	if !ok {
		handleError(w, http.StatusUnauthorized, "Deauth called without valid OAuth")
		return
	}

	oh.mutex.Lock()
	defer oh.mutex.Unlock()
	delete(oh.tokens, token)
	w.WriteHeader(http.StatusOK)
}

func (oh *OAuthHandler) isRequestValid(r *http.Request) (bool, int, string) {
	err := r.ParseForm()
	if err != nil {
		return false, http.StatusInternalServerError, fmt.Sprintf("Failed to parse form: %v", err)
	}

	clientID := r.FormValue(clientIDKey)
	clientSecret := r.FormValue(clientSecretKey)
	grantType := r.FormValue(grantTypeKey)

	if !oh.verifyClient(clientID, clientSecret) || grantType != "client_credentials" {
		return false, http.StatusForbidden, "Client verification failed"
	}

	return true, 0, ""
}

func (oh *OAuthHandler) verifyClient(id, secret string) bool {
	return id == oh.clientID && secret == oh.clientSecret
}
