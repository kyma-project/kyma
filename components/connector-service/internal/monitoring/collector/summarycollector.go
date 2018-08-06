package collector

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector interface {
	AddObservation(pathTemplate string, requestMethod string, message float64)
}

type summaryCollector struct {
	vector *prometheus.SummaryVec
}

func NewSummaryCollector(opts prometheus.SummaryOpts, labels []string) (Collector, apperrors.AppError) {
	vector := prometheus.NewSummaryVec(opts, labels)

	err := prometheus.Register(vector)
	if err != nil {
		return nil, apperrors.Internal("Failed to create metrics histogramCollector %s: %s", opts.Name, err.Error())
	}

	return &summaryCollector{vector: vector}, nil
}

func (ms *summaryCollector) AddObservation(pathTemplate string, requestMethod string, message float64) {
	ms.vector.WithLabelValues(pathTemplate, requestMethod).Observe(float64(message))
}