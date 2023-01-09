package subscribed

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy"
)

const responseName = "response"

// RespondWithBody sends http response with json body
func RespondWithBody(w http.ResponseWriter, events Events, httpCode int) {
	respond(w, httpCode)
	if err := json.NewEncoder(w).Encode(events); err != nil {
		namedLogger().Error(err)
	}
}

// RespondWithErrorAndLog logs error and sends http response with error json body
func RespondWithErrorAndLog(e error, w http.ResponseWriter) {
	namedLogger().Error(e.Error())
	respond(w, http.StatusInternalServerError)
	err := json.NewEncoder(w).Encode(legacy.HTTPErrorResponse{
		Code:  http.StatusInternalServerError,
		Error: e.Error(),
	})
	if err != nil {
		namedLogger().Error(err)
	}
}

func respond(w http.ResponseWriter, httpCode int) {
	w.Header().Set(internal.HeaderContentType, internal.ContentTypeApplicationJSON)
	w.WriteHeader(httpCode)
	namedLogger().Infof("Response code from \"subscribed\" request: HTTP %d", httpCode)
}

// Events represents collection of all events with subscriptions
type Events struct {
	EventsInfo []Event `json:"eventsInfo"`
}

// Event represents basic information about event
type Event struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func namedLogger() *zap.SugaredLogger {
	log, _ := logger.New("json", "info")
	return log.WithContext().Named(responseName)
}
