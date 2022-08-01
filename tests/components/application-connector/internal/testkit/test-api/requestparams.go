package test_api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type ExpectedRequestParameters struct {
	Headers         map[string][]string
	QueryParameters map[string][]string
}

func RequestParameters(expectedRequestParams ExpectedRequestParameters) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, expectedVals := range expectedRequestParams.Headers {
				actualVals := r.Header.Values(key)
				if !containsAllValues(actualVals, expectedVals) {
					handleError(w, http.StatusBadRequest, "Incorrect additional headers. Expected %s header to contain %v, but found %v", key, expectedVals, actualVals)
					return
				}
			}

			queryParameters := r.URL.Query()
			for key, expectedVals := range expectedRequestParams.QueryParameters {
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
