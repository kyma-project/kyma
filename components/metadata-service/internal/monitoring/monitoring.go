package monitoring

import (
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/monitoring/collector"
	"github.com/kyma-project/kyma/components/metadata-service/internal/monitoring/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	calculatedPercentile = 99.5
	percentileAccurace   = 0.05
)

func SetupMonitoringMiddleware() ([]mux.MiddlewareFunc, apperrors.AppError) {

	codeMiddleware, err := newCodeMiddleware("metadata_service_endpoints_responses")
	if err != nil {
		return nil, apperrors.Internal("Failed to setup code middleware: %s", err.Error())
	}

	durationMiddleware, err := newDurationMiddleware("metadata_service_endpoints_duration")
	if err != nil {
		return nil, apperrors.Internal("Failed to setup duration middleware: %s", err.Error())
	}

	return []mux.MiddlewareFunc{codeMiddleware.Handle, durationMiddleware.Handle}, nil
}

func newCodeMiddleware(name string) (middleware.Middleware, apperrors.AppError) {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: "help",
	}

	metricsCollector, err := collector.NewCounterCollector(opts, []string{"endpoint", "status", "method"})
	if err != nil {
		return nil, apperrors.Internal("Failed to setup response code metrics collector: %s", err.Error())
	}

	return middleware.NewCodeMiddleware(metricsCollector), nil
}

func newDurationMiddleware(name string) (middleware.Middleware, apperrors.AppError) {
	opts := prometheus.SummaryOpts{
		Name:       name,
		Help:       "help",
		Objectives: map[float64]float64{calculatedPercentile / 100: percentileAccurace},
	}

	metricsCollector, err := collector.NewSummaryCollector(opts, []string{"endpoint", "method"})
	if err != nil {
		return nil, apperrors.Internal("Failed to setup duration metrics collector: %s", err.Error())
	}

	return middleware.NewDurationMiddleware(metricsCollector), nil
}
