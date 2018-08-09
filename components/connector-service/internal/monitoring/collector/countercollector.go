package collector

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
)

type counterCollector struct {
	vector *prometheus.CounterVec
}

func NewCounterCollector(opts prometheus.CounterOpts, labels []string) (Collector, apperrors.AppError) {
	vector := prometheus.NewCounterVec(opts, labels)

	err := prometheus.Register(vector)
	if err != nil {
		return nil, apperrors.Internal("Failed to create counter collector %s: %s", opts.Name, err.Error())
	}

	return &counterCollector{vector: vector}, nil
}

func (ms *counterCollector) AddObservation(observation float64, labelValues ...string) {
	ms.vector.WithLabelValues(labelValues...).Add(float64(observation))
}
