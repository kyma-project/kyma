package collector

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
)

type counterCollector struct {
	vector *prometheus.CounterVec
}

func NewCounterCollector(opts prometheus.CounterOpts, labels []string) (Collector, apperrors.AppError) {
	vector := prometheus.NewCounterVec(opts, labels)

	err := prometheus.Register(vector)
	if err != nil {
		return nil, apperrors.Internal("Registering %s counter collector failed, %s", opts.Name, err.Error())
	}

	return &counterCollector{vector: vector}, nil
}

func (cc *counterCollector) AddObservation(observation float64, labelValues ...string) {
	cc.vector.WithLabelValues(labelValues...).Add(observation)
}
