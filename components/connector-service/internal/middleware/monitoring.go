package middleware

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/middleware/metrics"
	"net/http"
)

type Middleware interface {
	Handle(next http.Handler) http.Handler
}

func SetupMonitoring() ([]Middleware, apperrors.AppError) {

	durationMiddleware, err := newDurationMiddleware("connector_service_endpoints_duration")
	if err != nil {
		return nil, apperrors.Internal("Failed to setup duration middleware: %s", err.Error())
	}

	codeMiddleware, err := newCodeMiddleware("connector_service_endpoints_responses")
	if err != nil {
		return nil, apperrors.Internal("Failed to setup response codes middleware: %s", err.Error())
	}

	return []Middleware{durationMiddleware, codeMiddleware}, nil
}

func newCodeMiddleware(name string) (*codeMiddleware, apperrors.AppError) {
	metricsService, err := metrics.NewMetricsService(name, "Status codes returned by each endpoint", []string{"endpoint", "method"})
	if err != nil {
		return nil, apperrors.Internal("Failed to setup response codes metrics collector: %s", err.Error())
	}

	return NewCodeMiddleware(metricsService), nil
}

func newDurationMiddleware(name string) (*codeMiddleware, apperrors.AppError) {
	metricsService, err := metrics.NewMetricsService(name, "Response time for each endpoint", []string{"endpoint", "method"})
	if err != nil {
		return nil, apperrors.Internal("Failed to setup duration metrics collector: %s", err.Error())
	}

	return NewCodeMiddleware(metricsService), nil
}
