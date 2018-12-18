package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewHandler creates http.Handler(s) for the /v1/events and /v1/health endpoints
func NewHandler() http.Handler {
	router := mux.NewRouter()

	router.PathPrefix("/{re}/v1/events").Handler(NewEventsHandler()).Methods(http.MethodPost)

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}
