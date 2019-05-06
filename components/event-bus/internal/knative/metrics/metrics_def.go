package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace        = "namespace"
	Name             = "name"
	Ready            = "ready"
	//Subscriber		 = "subscriber"
)

type SubscriptionsGauge struct {
	Labels []string
	Metric *prometheus.GaugeVec
}

var (
	KymaSubscriptionsGaugeObj *SubscriptionsGauge
	kymaSubscriptionsGaugeVec *prometheus.GaugeVec
	kymaSubscriptionsGaugeLabels = []string{Namespace, Name, Ready}

	KnativeSubscriptionsGaugeObj *SubscriptionsGauge
	knativeSubscriptionsGaugeVec *prometheus.GaugeVec
	knativeSubscriptionsGaugeLabels = []string{Name, Ready}

	KnativeChanelGaugeObj *SubscriptionsGauge
	knativeChanelGaugeVec *prometheus.GaugeVec
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

func (ksg *SubscriptionsGauge) DeleteKymaSubscriptionsGauge(namespace string, name string) {
	values := []string{namespace, name, "true"}
	ksg.Metric.DeleteLabelValues(values...)
	values = []string{namespace, name, "false"}
	ksg.Metric.DeleteLabelValues(values...)
}

func (ksg *SubscriptionsGauge) DeleteKnativeSubscriptionsGauge(name string) {
	ksg.Metric.DeleteLabelValues(name, "true")
	ksg.Metric.DeleteLabelValues(name, "false")
}


func (kchg *SubscriptionsGauge) DeleteKnativeChannelGauge(name string) {
	kchg.Metric.DeleteLabelValues(name)
}
