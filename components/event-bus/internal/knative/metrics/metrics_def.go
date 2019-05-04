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

var (
	KymaSubscriptionsGaugeObj *KymaSubscriptionsGauge
	kymaSubscriptionsGaugeVec *prometheus.GaugeVec
	kymaSubscriptionsGaugeLabels = []string{Namespace, Name, Ready}
)

func init() {
	kymaSubscriptionsGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_kyma_subscriptions",
		Help: "The total number of Kyma subscriptions",
	}, kymaSubscriptionsGaugeLabels)
	KymaSubscriptionsGaugeObj = &KymaSubscriptionsGauge{
		Labels: kymaSubscriptionsGaugeLabels,
		Metric: kymaSubscriptionsGaugeVec,
	}
}

func (ksg *KymaSubscriptionsGauge) DeleteKymaSubscriptionsGauge(namespace string, name string) {
	values := []string{namespace, name, "true"}
	ksg.Metric.DeleteLabelValues(values...)
	values = []string{namespace, name, "false"}
	ksg.Metric.DeleteLabelValues(values...)
}
