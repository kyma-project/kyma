package httperrors

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"net/http"
)

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func AppErrorToResponse(appError apperrors.AppError, detailedErrorResponse bool) (status int, body ErrorResponse) {
	httpCode := errorCodeToHttpStatus(appError.Code())
	errorMessage := appError.Error()
	return formatErrorResponse(httpCode, errorMessage, detailedErrorResponse)
}

func errorCodeToHttpStatus(code int) int {
	switch code {
	case apperrors.CodeInternal:
		return http.StatusInternalServerError
	case apperrors.CodeNotFound:
		return http.StatusNotFound
	case apperrors.CodeAlreadyExists:
		return http.StatusConflict
	case apperrors.CodeWrongInput:
		return http.StatusBadRequest
	case apperrors.CodeUpstreamServerCallFailed:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func formatErrorResponse(httpCode int, errorMessage string, detailedErrorResponse bool) (status int, body ErrorResponse) {
	if isInternalError(httpCode) && !detailedErrorResponse {
		return httpCode, ErrorResponse{httpCode, "Internal error."}
	}
	return httpCode, ErrorResponse{httpCode, errorMessage}
}

func isInternalError(httpCode int) bool {
	return httpCode == http.StatusInternalServerError
}
