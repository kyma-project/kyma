package test_api

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
)

func AddAPIHandler(router *mux.Router, oauthTokesCache map[string]bool, csrfTokensCache map[string]bool, mutex *sync.RWMutex, basicAuthCredentials BasicAuthCredentials) {

	alwaysOKHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	echoHandler := NewEchoHandler()

	router.HandleFunc("/v1/api/unsecure/ok", alwaysOKHandler).Methods("GET", "PUT", "POST")
	router.HandleFunc("/v1/api/unsecure/echo", echoHandler).Methods("GET", "PUT", "POST")
	router.HandleFunc("/v1/api/basic/ok", NewBasicAuthHandler(basicAuthCredentials, alwaysOKHandler)).Methods("GET", "PUT", "POST")
	router.HandleFunc("/v1/api/basic/echo", NewBasicAuthHandler(basicAuthCredentials, echoHandler)).Methods("GET", "PUT", "POST")
	router.HandleFunc("/v1/api/oauth/ok", NewOAuthHandler(oauthTokesCache, mutex, alwaysOKHandler)).Methods("GET", "PUT", "POST")
	router.HandleFunc("/v1/api/oauth/echo", NewOAuthHandler(oauthTokesCache, mutex, echoHandler)).Methods("GET", "PUT", "POST")
	router.HandleFunc("/v1/health", alwaysOKHandler).Methods("GET")
}

func NewOAuthHandler(oauthTokesCache map[string]bool, mutex *sync.RWMutex, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		mutex.RLock()
		_, found := oauthTokesCache[token]
		mutex.RUnlock()

		if !found {
			handleError(w, http.StatusForbidden, "Invalid token")
			return
		}

		next.ServeHTTP(w, r)
	}
}

type BasicAuthCredentials struct {
	User     string
	Password string
}

func NewBasicAuthHandler(credentials BasicAuthCredentials, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok {
			handleError(w, http.StatusForbidden, "Basic auth header not found")
			return
		}

		if user != credentials.User || password != credentials.Password {
			handleError(w, http.StatusForbidden, "Incorrect username or Password")
			return
		}

		next.ServeHTTP(w, r)
	}
}

func handleError(w http.ResponseWriter, code int, format string, a ...interface{}) {
	err := errors.New(fmt.Sprintf(format, a...))
	log.Error(err)
	w.WriteHeader(code)
}
