package test_api

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

const (
	csrfTokenHeader = "X-csrf-token"
	csrfTokenCookie = "csrftokencookie"
)

type CSRFTokens map[string]interface{}

type CSRFHandler struct {
	mutex  sync.RWMutex
	tokens map[string]interface{}
}

func NewCSRF(tokens CSRFTokens) CSRFHandler {
	return CSRFHandler{
		mutex:  sync.RWMutex{},
		tokens: tokens,
	}
}

func (ch *CSRFHandler) Token(w http.ResponseWriter, _ *http.Request) {
	token := uuid.New().String()

	ch.mutex.Lock()
	ch.tokens[token] = nil
	ch.mutex.Unlock()

	w.Header().Set(csrfTokenHeader, token)
	http.SetCookie(w, &http.Cookie{
		Name:  csrfTokenCookie,
		Value: token,
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}

func (ch *CSRFHandler) BadToken(w http.ResponseWriter, _ *http.Request) {
	token := uuid.New().String()

	w.Header().Set(csrfTokenHeader, token)
	http.SetCookie(w, &http.Cookie{
		Name:  csrfTokenCookie,
		Value: token,
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}

func (ch *CSRFHandler) Middleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerToken := r.Header.Get(csrfTokenHeader)
			if headerToken == "" {
				handleError(w, http.StatusForbidden, "CSRF token header missing")
				return
			}

			ch.mutex.RLock()
			_, found := ch.tokens[headerToken]
			ch.mutex.RUnlock()

			if !found {
				handleError(w, http.StatusForbidden, "Invalid CSRF token from the header")
				return
			}

			cookieToken, err := r.Cookie(csrfTokenCookie)
			if err != nil {
				handleError(w, http.StatusForbidden, "CSRF token cookie missing")
				return
			}

			ch.mutex.RLock()
			_, found = ch.tokens[cookieToken.Value]
			ch.mutex.RUnlock()

			if !found {
				handleError(w, http.StatusForbidden, "Invalid CSRF token from the cookie")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
