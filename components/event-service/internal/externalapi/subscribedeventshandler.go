package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
)

type activeEventsHandler struct {
	EventsClient subscribed.EventsClient
}

//NewActiveEventsHandler creates handler to handle subscribedevents endpoint
func NewActiveEventsHandler(eventsClient subscribed.EventsClient) SubscribedEventsHandler {
	return &activeEventsHandler{EventsClient: eventsClient}
}

func (aeh *activeEventsHandler) GetSubscribedEvents(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["application"]

	// TODO(marcobebway) get knative triggers and return a set of unique trigger.filters...events
	events, e := aeh.EventsClient.GetSubscribedEvents(appName)

	if e != nil {
		httptools.RespondWithErrorAndLog(e, w)
		return
	}

	httptools.RespondWithBody(w, events, http.StatusOK)
}
