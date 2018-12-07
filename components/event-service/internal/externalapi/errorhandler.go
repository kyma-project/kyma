package externalapi

import (
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/event-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/event-service/internal/httperrors"
)

// An ErrorHandler represents an error with a message and a status code
type ErrorHandler struct {
	Message string
	Code    int
}

// NewErrorHandler creates a new ErrorHandler with the given code and message
func NewErrorHandler(code int, message string) *ErrorHandler {
	return &ErrorHandler{Message: message, Code: code}
}

func (eh *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseBody := httperrors.ErrorResponse{Code: eh.Code, Error: eh.Message}

	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJSONWithCharset)
	w.WriteHeader(eh.Code)
	json.NewEncoder(w).Encode(responseBody)
}
