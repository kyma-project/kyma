package mock

import (
	"encoding/json"
	"net/http"
)

type ErrorHandler struct {
	Message string
	Code    int
}

type ErrorResponse struct {
	Code  int
	Error string
}

func NewErrorHandler(code int, message string) *ErrorHandler {
	return &ErrorHandler{Message: message, Code: code}
}

func (eh *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseBody := ErrorResponse{Code: eh.Code, Error: eh.Message}

	w.WriteHeader(eh.Code)
	json.NewEncoder(w).Encode(responseBody)
}
