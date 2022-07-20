package test_api

import (
	"io"
	"net/http"
	"reflect"

	"github.com/go-http-utils/logger"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
			//if len(r.Header) != len(expectedHeaders) {
			//	handleError(w, http.StatusForbidden, "Incorrect username or Password")
			//	return
			//}
			//for key, vals := range expectedHeaders {
			//	v := r.Header.Values(key)
			//	if !reflect.DeepEqual(vals, v) {
			//		handleError(w, http.StatusForbidden, "Incorrect username or Password")
			//		return
			//	}
			//}
			//
			//queryParameters := r.URL.Query()
			//if len(queryParameters) != len(expectedQueryParameters) {
			//	handleError(w, http.StatusForbidden, "Incorrect username or Password")
			//	return
			//}
			//for key, vals := range expectedQueryParameters {
			//	v := queryParameters[key]
			//	if !reflect.DeepEqual(vals, v) {
			//		handleError(w, http.StatusForbidden, "Incorrect username or Password")
			//		return
			//	}
			//}
			log.Warnf("r.Header: %v\nexpectedHeaders: %v", r.Header, expectedHeaders)
			log.Warnf("r.URL.Query(): %v\nexpectedQueryParameters: %v", r.URL.Query(), expectedQueryParameters)
			if !reflect.DeepEqual(r.Header, expectedHeaders) || !reflect.DeepEqual(r.URL.Query(), expectedQueryParameters) {
				handleError(w, http.StatusBadRequest, "Incorrect headers or query parameters")
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
