package middleware

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"net/http"
)

type Middleware interface {
	Handle(next http.Handler) http.Handler
}

func SetupMonitoring() ([]Middleware, apperrors.AppError) {

	durationMiddleware, err := NewDurationMiddleware("connector_service_endpoints_duration")
	if err != nil {
		return nil, apperrors.Internal("Failed to setup duration middleware: %s", err.Error())
	}

	codeMiddleware, err := NewCodeMiddleware("connector_service_endpoints_responses")
	if err != nil {
		return nil, apperrors.Internal("Failed to setup response codes middleware: %s", err.Error())
	}

	return []Middleware{durationMiddleware, codeMiddleware}, nil
}
