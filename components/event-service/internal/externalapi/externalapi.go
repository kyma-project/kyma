package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"
)

//SubscribedEventsHandler is an interface representing handler for the /v1/subscribedevents endpoint
type SubscribedEventsHandler interface {
	GetSubscribedEvents(w http.ResponseWriter, r *http.Request)
}

// NewHandler creates http.Handler(s) for the /v1/events /v2/events /v1/events/subscribed and /v1/health endpoints
func NewHandler(maxRequestSize int64, eventsClient subscribed.EventsClient, eventMeshURL string) http.Handler {
	router := mux.NewRouter()

	router.Path("/{application}/v1/events").Handler(NewEventsHandler(maxRequestSize)).Methods(http.MethodPost)

	// v2 endpoint is moved permanently
	router.Path("/{application}/v2/events").Handler(NewPermanentRedirectionHandler(eventMeshURL)).Methods(http.MethodPost)

	// TODO(marcobebway) respect this contract
	router.Path("/{application}/v1/events/subscribed").HandlerFunc(NewActiveEventsHandler(eventsClient).GetSubscribedEvents).Methods(http.MethodGet)

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.NotFoundHandler = NewErrorHandler(http.StatusNotFound, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(http.StatusMethodNotAllowed, "Method not allowed.")

	return router
}
