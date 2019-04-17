package externalapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/event-service/internal/events/registered"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
)

type activeEventsHandler struct {
	EventsClient registered.EventsClient
}

//NewActiveEventsHandler creates handler to handle activeevents endpoint
func NewActiveEventsHandler(eventsClient registered.EventsClient) ActiveEventsHandler {
	return &activeEventsHandler{EventsClient: eventsClient}
}

func (aeh *activeEventsHandler) GetActiveEvents(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["application"]

	events, e := aeh.EventsClient.GetActiveEvents(appName)

	if e != nil {
		httptools.RespondWithErrorAndLog(e, w)
		return
	}

	httptools.RespondWithBody(w, events, http.StatusOK)
}
