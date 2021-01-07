package subscribed

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/sirupsen/logrus"

	"encoding/json"
)

//RespondWithBody sends http response with json body
func RespondWithBody(w http.ResponseWriter, events Events, httpCode int) {
	respond(w, httpCode)
	json.NewEncoder(w).Encode(events)
}

//RespondWithErrorAndLog logs error and sends http response with error json body
func RespondWithErrorAndLog(e error, w http.ResponseWriter) {
	logrus.Errorln(e.Error())
	respond(w, http.StatusInternalServerError)
	json.NewEncoder(w).Encode(legacy.HTTPErrorResponse{
		Code:  http.StatusInternalServerError,
		Error: e.Error(),
	})
}

func respond(w http.ResponseWriter, httpCode int) {
	w.Header().Set(legacy.HeaderContentType, legacy.ContentTypeApplicationJSON)
	w.WriteHeader(httpCode)
	logrus.Infof("Response code: HTTP %d", httpCode)
}

//Events represents collection of all events with subscriptions
type Events struct {
	EventsInfo []Event `json:"eventsInfo"`
}

//Event represents basic information about event
type Event struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
