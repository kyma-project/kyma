package httptools

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/pkg/logger"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/httperrors"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/apperrors"
)

func RespondWithError(log *logger.Logger, w http.ResponseWriter, apperr apperrors.AppError) {
	log.Error(apperr.Error())

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
