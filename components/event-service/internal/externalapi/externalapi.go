package externalapi

import (
	"net/http"

	v1 "github.com/kyma-project/kyma/components/event-service/internal/externalapi/v1"
	v2 "github.com/kyma-project/kyma/components/event-service/internal/externalapi/v2"

	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"

	"github.com/gorilla/mux"

	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventstransport "github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

//SubscribedEventsHandler is an interface representing handler for the /v1/subscribedevents endpoint
type SubscribedEventsHandler interface {
	GetSubscribedEvents(w http.ResponseWriter, r *http.Request)
}

// NewHandler creates http.Handler(s) for the /v1/events /v2/events /v1/subscribedevents and /v1/health endpoints
func NewHandler(maxRequestSize int64, eventsClient subscribed.EventsClient) http.Handler {
	router := mux.NewRouter()

	router.Path("/{application}/v1/events").Handler(v1.NewEventsHandler(maxRequestSize)).Methods(http.MethodPost)

	t, err := cloudevents.NewHTTPTransport()
	if err != nil {
		return nil
	}

	//TODO: set the logic here
	t.SetReceiver(cloudeventstransport.ReceiveFunc(v2.HandleEvent))

	router.Path("/{application}/v2/events").Handler(t).Methods(http.MethodPost)

	router.Path("/{application}/v1/events/subscribed").HandlerFunc(NewActiveEventsHandler(eventsClient).GetSubscribedEvents).Methods(http.MethodGet)

	router.Path("/v1/health").Handler(NewHealthCheckHandler()).Methods(http.MethodGet)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	return router
}
