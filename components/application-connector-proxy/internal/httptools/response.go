package httptools

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/application-connector-proxy/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-connector-proxy/internal/httperrors"
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector-proxy/internal/apperrors"
	log "github.com/sirupsen/logrus"
)

func RespondWithErrorAndLog(w http.ResponseWriter, apperr apperrors.AppError) {
	log.Errorln(apperr.Error())

	RespondWithError(w, apperr)
}

func RespondWithError(w http.ResponseWriter, apperr apperrors.AppError) {
	statusCode, responseBody := httperrors.AppErrorToResponse(apperr)

	Respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func Respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func RespondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	Respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}
