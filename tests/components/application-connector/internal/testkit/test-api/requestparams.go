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
				if !containsSubset(actualVals, expectedVals) {
					handleError(w, http.StatusBadRequest, "Incorrect additional headers. Expected %s header to contain %v, but found %v", key, expectedVals, actualVals)
					return
				}
			}

			queryParameters := r.URL.Query()
			for key, expectedVals := range expectedRequestParams.QueryParameters {
				actualVals := queryParameters[key]
				if !containsSubset(actualVals, expectedVals) {
					handleError(w, http.StatusBadRequest, "Incorrect additional query parameters. Expected %s query parameter to contain %v, but found %v", key, expectedVals, actualVals)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func containsSubset(set, subset []string) bool {
	for _, bVal := range subset {
		found := false
		for _, aVal := range set {
			if aVal == bVal {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	return true
}
