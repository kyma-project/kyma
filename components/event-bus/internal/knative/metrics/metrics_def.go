package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// Namespace label
	Namespace = "namespace"
	// Name label
	Name = "name"
	// Ready label
	Ready = "ready"
)

// SubscriptionsGauge represents the Subscriptions Gauge.
type SubscriptionsGauge struct {
	Labels []string
	Metric *prometheus.GaugeVec
}

var (
	// KymaSubscriptionsGaugeObj instance
	KymaSubscriptionsGaugeObj *SubscriptionsGauge
	// kymaSubscriptionsGaugeVec instance
	kymaSubscriptionsGaugeVec *prometheus.GaugeVec
	// kymaSubscriptionsGaugeLabels instance
	kymaSubscriptionsGaugeLabels = []string{Namespace, Name, Ready}

	// KnativeSubscriptionsGaugeObj instance
	KnativeSubscriptionsGaugeObj *SubscriptionsGauge
	// knativeSubscriptionsGaugeVec instance
	knativeSubscriptionsGaugeVec *prometheus.GaugeVec
	// knativeSubscriptionsGaugeLabels instance
	knativeSubscriptionsGaugeLabels = []string{Namespace, Name, Ready}

	// KnativeChanelGaugeObj instance
	KnativeChanelGaugeObj *SubscriptionsGauge
	// knativeChanelGaugeVec instance
	knativeChanelGaugeVec *prometheus.GaugeVec
	// knativeChanelGaugeLabels instance
	knativeChanelGaugeLabels = []string{Name}
)

func init() {
	kymaSubscriptionsGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_kyma_subscriptions",
		Help: "The total number of Kyma subscriptions",
	}, kymaSubscriptionsGaugeLabels)
	KymaSubscriptionsGaugeObj = &SubscriptionsGauge{
		Labels: kymaSubscriptionsGaugeLabels,
		Metric: kymaSubscriptionsGaugeVec,
	}

	knativeSubscriptionsGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_knative_subscriptions",
		Help: "The total number of Knative subscriptions",
	}, knativeSubscriptionsGaugeLabels)
	KnativeSubscriptionsGaugeObj = &SubscriptionsGauge{
		Labels: knativeSubscriptionsGaugeLabels,
		Metric: knativeSubscriptionsGaugeVec,
	}

	knativeChanelGaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_knative_channels",
		Help: "The total number of Knative channels",
	}, knativeChanelGaugeLabels)
	KnativeChanelGaugeObj = &SubscriptionsGauge{
		Labels: knativeChanelGaugeLabels,
		Metric: knativeChanelGaugeVec,
	}
}

// DeleteKymaSubscriptionsGaugeLabelValues deletes the Kyma subscriptions gauge label values.
func (ksg *SubscriptionsGauge) DeleteKymaSubscriptionsGaugeLabelValues(namespace string, name string) {
	values := []string{namespace, name, "true"}
	ksg.Metric.DeleteLabelValues(values...)
	values = []string{namespace, name, "false"}
	ksg.Metric.DeleteLabelValues(values...)
}

// DeleteKnativeSubscriptionsGaugeLabelValues deletes the Knative subscriptions gauge label values.
func (ksg *SubscriptionsGauge) DeleteKnativeSubscriptionsGaugeLabelValues(namespace string, name string) {
	ksg.Metric.DeleteLabelValues(namespace, name, "true")
	ksg.Metric.DeleteLabelValues(namespace, name, "false")
}

// DeleteKnativeChannelGaugeLabelValues deletes the Knative channel gauge label values.
func (ksg *SubscriptionsGauge) DeleteKnativeChannelGaugeLabelValues(name string) {
	ksg.Metric.DeleteLabelValues(name)
}
