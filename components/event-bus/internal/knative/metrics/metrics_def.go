package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace        = "namespace"
	Name             = "name"
	Status           = "status"
	SourceID         = "source_id"
	EventType        = "event_type"
	EventTypeVersion = "event_type_version"
	Ready            = "ready"
	EventsActivated  = "events_activated"
)

type KymaSubscriptionsGauge struct {
	Labels []string
	Metric *prometheus.GaugeVec
}

func NewKymaSubscriptionsGauge() *KymaSubscriptionsGauge {
	labels := []string{Namespace, Name, Ready}
	kymaSubscriptionsGauge := KymaSubscriptionsGauge{
		Labels: labels,
		Metric: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "total_kyma_subscriptions",
			Help: "The total number of Kyma subscriptions",
		}, labels),
	}
	return &kymaSubscriptionsGauge
}

func (ksg *KymaSubscriptionsGauge) DeleteKymaSubscriptionsGauge(values []string) {
	ksg.Metric.DeleteLabelValues(values...)
}
