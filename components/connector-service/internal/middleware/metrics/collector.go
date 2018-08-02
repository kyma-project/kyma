package metrics

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector interface {
	AddObservation(pathTemplate string, requestMethod string, message float64)
}

type collector struct {
	vector *prometheus.SummaryVec
}

func NewMetricsService(name string, help string, labels []string) (Collector, apperrors.AppError) {
	vector := newSummaryVec(name, help, labels)

	err := prometheus.Register(vector)
	if err != nil {
		return nil, apperrors.Internal("Failed to create metrics collector %s: %s", name, err.Error())
	}

	return &collector{vector: vector}, nil
}

func newSummaryVec(name string, help string, labels []string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
}

func (ms *collector) AddObservation(pathTemplate string, requestMethod string, message float64) {
	ms.vector.WithLabelValues(pathTemplate, requestMethod).Observe(float64(message))
}
