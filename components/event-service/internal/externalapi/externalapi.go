package externalapi

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"
	v2 "github.com/kyma-project/kyma/components/event-service/internal/externalapi/v2"

	"github.com/gorilla/mux"
)

//SubscribedEventsHandler is an interface representing handler for the /v1/subscribedevents endpoint
type SubscribedEventsHandler interface {
	GetSubscribedEvents(w http.ResponseWriter, r *http.Request)
}

// NewHandler creates http.Handler(s) for the /v1/events /v2/events /v1/subscribedevents and /v1/health endpoints
func NewHandler(maxRequestSize int64, eventsClient subscribed.EventsClient) http.Handler {
	router := mux.NewRouter()

	router.Path("/{application}/v1/events").Handler(NewEventsHandler(maxRequestSize)).Methods(http.MethodPost)

	// TODO(marcobebway) cleanup the following
	//router.Path("/{application}/v1/events").Handler(v1.NewEventsHandler(maxRequestSize)).Methods(http.MethodPost)

	// TODO(marcobebway) return 3xx for redirect
	router.Path("/{application}/v2/events").Handler(v2.NewEventsHandler(maxRequestSize)).Methods(http.MethodPost)

	// TODO(marcobebway) respect this contract
	router.Path("/{application}/v1/events/subscribed").HandlerFunc(NewActiveEventsHandler(eventsClient).GetSubscribedEvents).Methods(http.MethodGet)

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}
