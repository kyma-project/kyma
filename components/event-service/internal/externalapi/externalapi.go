package externalapi

import (
	"github.com/kyma-project/kyma/components/event-service/internal/events/registered"
	"net/http"

	"github.com/gorilla/mux"
)

type ActiveEventsHandler interface {
	GetActiveEvents(w http.ResponseWriter, r *http.Request)
}

// NewHandler creates http.Handler(s) for the /v1/events and /v1/health endpoints
func NewHandler(maxRequestSize int64) http.Handler {
	router := mux.NewRouter()

	router.PathPrefix("/{re}/v1/events").Handler(NewEventsHandler(maxRequestSize)).Methods(http.MethodPost)

	eventsClient, _ := registered.NewEventsClient()
	router.Path("{application}/v1/activeevents").HandlerFunc(NewActiveEventsHandler(eventsClient).GetActiveEvents).Methods(http.MethodGet)

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}
