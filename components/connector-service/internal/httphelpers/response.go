package httphelpers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
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
