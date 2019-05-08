package externalapi

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"

	"github.com/gorilla/mux"
)

//SubscribedEventsHandler is an interface representing handler for the /v1/subscribedevents endpoint
type SubscribedEventsHandler interface {
	GetSubscribedEvents(w http.ResponseWriter, r *http.Request)
}

// NewHandler creates http.Handler(s) for the /v1/events /v1/subscribedevents and /v1/health endpoints
func NewHandler(maxRequestSize int64, eventsClient subscribed.EventsClient) http.Handler {
	router := mux.NewRouter()

	router.Path("/{application}/v1/events").Handler(NewEventsHandler(maxRequestSize)).Methods(http.MethodPost)

	router.Path("/{application}/v1/events/subscribed").HandlerFunc(NewActiveEventsHandler(eventsClient).GetSubscribedEvents).Methods(http.MethodGet)

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}
