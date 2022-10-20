package test_api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type BasicAuthCredentials struct {
	User     string
	Password string
}

func BasicAuth(credentials BasicAuthCredentials) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok {
				handleError(w, http.StatusForbidden, "Basic auth header not found")
				return
			}

			if credentials.User != u || credentials.Password != p {
				handleError(w, http.StatusForbidden, "Incorrect username or Password")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
