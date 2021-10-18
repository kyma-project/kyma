package collector

import (
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector interface {
	AddObservation(observation float64, labelValues ...string)
}

type summaryCollector struct {
	vector *prometheus.SummaryVec
}

func NewSummaryCollector(opts prometheus.SummaryOpts, labels []string) (Collector, apperrors.AppError) {
	vector := prometheus.NewSummaryVec(opts, labels)

	err := prometheus.Register(vector)
	if err != nil {
		return nil, apperrors.Internal("Failed to create summary collector %s: %s", opts.Name, err.Error())
	}

	return &summaryCollector{vector: vector}, nil
}

func (ms *summaryCollector) AddObservation(observation float64, labelValues ...string) {
	ms.vector.WithLabelValues(labelValues...).Observe(float64(observation))
}
