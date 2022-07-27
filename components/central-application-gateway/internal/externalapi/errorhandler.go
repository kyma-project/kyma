package externalapi

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/httperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
)

type ErrorHandler struct {
	Message string
	Code    int
}

func NewErrorHandler(code int, message string) *ErrorHandler {
	return &ErrorHandler{Message: message, Code: code}
}

func (eh *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseBody := httperrors.ErrorResponse{Code: eh.Code, Error: eh.Message}

	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(eh.Code)
	_ = json.NewEncoder(w).Encode(responseBody)
}
