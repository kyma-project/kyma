package monitoring

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring/collector"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

const (
	calculatedPercentile = 99.5
	percentileAccuracy   = 0.05
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

func newCodeMiddleware(name string) (*middleware.CodeMiddleware, apperrors.AppError) {
	opts := prometheus.CounterOpts{
		Name: name,
		Help: "help",
	}

	metricsCollector, err := collector.NewCounterCollector(opts, []string{"endpoint", "status", "method"})
	if err != nil {
		return nil, apperrors.Internal("Failed to setup response codes metrics collector: %s", err.Error())
	}

	return middleware.NewCodeMiddleware(metricsCollector), nil
}

func newDurationMiddleware(name string) (*middleware.DurationMiddleware, apperrors.AppError) {
	opts := prometheus.SummaryOpts{
		Name:       name,
		Help:       "help",
		Objectives: map[float64]float64{calculatedPercentile / 100: percentileAccuracy},
	}

	metricsCollector, err := collector.NewSummaryCollector(opts, []string{"endpoint", "method"})
	if err != nil {
		return nil, apperrors.Internal("Failed to setup duration metrics collector: %s", err.Error())
	}

	return middleware.NewDurationMiddleware(metricsCollector), nil
}
