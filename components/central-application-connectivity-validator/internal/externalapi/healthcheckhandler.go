package externalapi

import (
	"net/http"
)

// NewHealthCheckHandler creates handler for performing health check
func NewHealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
