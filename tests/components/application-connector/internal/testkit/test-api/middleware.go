package test_api

import (
	"io"
	"net/http"

	"github.com/go-http-utils/logger"
	"github.com/gorilla/mux"
)

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

func RequestParameters(expectedHeaders, expectedQueryParameters map[string][]string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, expectedVals := range expectedHeaders {
				actualVals := r.Header.Values(key)
				if !containsAllValues(actualVals, expectedVals) {
					handleError(w, http.StatusBadRequest, "Incorrect additional headers. Expected %s header to contain %v, but found %v", key, expectedVals, actualVals)
					return
				}
			}

			queryParameters := r.URL.Query()
			for key, expectedVals := range expectedQueryParameters {
				actualVals := queryParameters[key]
				if !containsAllValues(actualVals, expectedVals) {
					handleError(w, http.StatusBadRequest, "Incorrect additional query parameters. Expected %s query parameter to contain %v, but found %v", key, expectedVals, actualVals)
					return
				}
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

func containsAllValues[T comparable](a, b []T) bool {
	for _, bVal := range b {
		for _, aVal := range a {
			if bVal == aVal {
				return true
			}
		}
	}
	return false
}
