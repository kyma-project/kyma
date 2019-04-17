package httptools

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/event-service/internal/events/registered"
	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/event-service/internal/httperrors"
	"github.com/sirupsen/logrus"
	"net/http"
)

func RespondWithBody(w http.ResponseWriter, events registered.ActiveEvents, httpCode int) {
	Respond(w, httpCode)
	json.NewEncoder(w).Encode(events)
}

func RespondWithErrorAndLog(e error, w http.ResponseWriter) {
	logrus.Errorln(e.Error())
	Respond(w, http.StatusInternalServerError)
	json.NewEncoder(w).Encode(httperrors.ErrorResponse{
		Code:  http.StatusInternalServerError,
		Error: e.Error(),
	})
}

func Respond(w http.ResponseWriter, httpCode int) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJSON)
	w.WriteHeader(httpCode)
}
