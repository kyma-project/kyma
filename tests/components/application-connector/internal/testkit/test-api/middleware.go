package test_api

import (
	"io"
	"net/http"

	"github.com/go-http-utils/logger"
	"github.com/gorilla/mux"
)

type ContextKey int

func BasicAuth(user, pass string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok {
				handleError(w, http.StatusForbidden, "Basic auth header not found")
				return
			}

			if user != u || pass != p {
				handleError(w, http.StatusForbidden, "Incorrect username or Password")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func Logger(out io.Writer, t logger.Type) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return logger.Handler(next, out, t)
	}
}
