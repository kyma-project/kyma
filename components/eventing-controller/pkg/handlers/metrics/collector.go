package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// deliveryMetricKey name of the delivery per subscription metric
	deliveryMetricKey = "delivery_per_subscription"

	// deliveryMetricHelp help text for the delivery_per_subscription metric
	deliveryMetricHelp = "Number of dispatched events per subscription"
)

// Collector implements the prometheus.Collector interface
type Collector struct {
	deliveryPerSubscription *prometheus.CounterVec
}

// NewCollector a new instance of Collector
func NewCollector() *Collector {
	return &Collector{
		deliveryPerSubscription: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: deliveryMetricKey,
				Help: deliveryMetricHelp,
			},
			[]string{"subscription_name", "event_type", "sink", "response_code"},
		),
	}
}

// Describe implements the prometheus.Collector interface Describe method
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.deliveryPerSubscription.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.deliveryPerSubscription.Collect(ch)
}

// RecordDeliveryPerSubscription records the delivery_per_subscription metric
func (c *Collector) RecordDeliveryPerSubscription(subscriptionName, eventType, sink string, statusCode int) {
	c.deliveryPerSubscription.WithLabelValues(subscriptionName, eventType, fmt.Sprintf("%v", sink), fmt.Sprintf("%v", statusCode)).Inc()
}

// RegisterMetrics registers the metrics
func (c *Collector) RegisterMetrics() {
	metrics.Registry.MustRegister(c.deliveryPerSubscription)
}
