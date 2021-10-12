package httptools

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/httpconsts"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/httperrors"

	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/apperrors"
)

func RespondWithError(log *zap.SugaredLogger, w http.ResponseWriter, apperr apperrors.AppError) {
	log.Errorf("Error: %s", apperr.Error())

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
