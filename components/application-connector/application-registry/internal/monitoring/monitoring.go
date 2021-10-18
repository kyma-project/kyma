package monitoring

import (
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/monitoring/collector"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/monitoring/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	calculatedPercentile = 99.5
	percentileAccuracy   = 0.05
)

func SetupMonitoringMiddleware() ([]mux.MiddlewareFunc, apperrors.AppError) {

	codeMiddleware, err := newCodeMiddleware("application_registry_endpoints_responses")
	if err != nil {
		return nil, apperrors.Internal("Setup code middleware failed, %s", err.Error())
	}

	durationMiddleware, err := newDurationMiddleware("application_registry_endpoints_duration")
	if err != nil {
		return nil, apperrors.Internal("Setup duration middleware failed, %s", err.Error())
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
		return nil, apperrors.Internal("Setup response code metrics collector failed, %s", err.Error())
	}

	return middleware.NewCodeMiddleware(metricsCollector), nil
}

func newDurationMiddleware(name string) (middleware.Middleware, apperrors.AppError) {
	opts := prometheus.SummaryOpts{
		Name:       name,
		Help:       "help",
		Objectives: map[float64]float64{calculatedPercentile / 100: percentileAccuracy},
	}

	metricsCollector, err := collector.NewSummaryCollector(opts, []string{"endpoint", "method"})
	if err != nil {
		return nil, apperrors.Internal("Setup response duration metrics collector failed, %s", err.Error())
	}

	return middleware.NewDurationMiddleware(metricsCollector), nil
}
