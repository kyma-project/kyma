package httptools

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/event-service/internal/events/registered"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/event-service/internal/httperrors"
	"github.com/sirupsen/logrus"
	"net/http"
)

//RespondWithBody sends http response with json body
func RespondWithBody(w http.ResponseWriter, events registered.ActiveEvents, httpCode int) {
	respond(w, httpCode)
	json.NewEncoder(w).Encode(events)
}

//RespondWithErrorAndLog logs error and sends http response with error json body
func RespondWithErrorAndLog(e error, w http.ResponseWriter) {
	logrus.Errorln(e.Error())
	respond(w, http.StatusInternalServerError)
	json.NewEncoder(w).Encode(httperrors.ErrorResponse{
		Code:  http.StatusInternalServerError,
		Error: e.Error(),
	})
}

func respond(w http.ResponseWriter, httpCode int) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJSON)
	w.WriteHeader(httpCode)
}
