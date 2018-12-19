package collector

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/prometheus/client_golang/prometheus"
)

type summaryCollector struct {
	vector *prometheus.SummaryVec
}

func NewSummaryCollector(opts prometheus.SummaryOpts, labels []string) (Collector, apperrors.AppError) {
	vector := prometheus.NewSummaryVec(opts, labels)

	err := prometheus.Register(vector)
	if err != nil {
		return nil, apperrors.Internal("Registering %s summary collector failed, %s", opts.Name, err.Error())
	}

	return &summaryCollector{vector: vector}, nil
}

func (sc *summaryCollector) AddObservation(observation float64, labelValues ...string) {
	sc.vector.WithLabelValues(labelValues...).Observe(observation)
}
