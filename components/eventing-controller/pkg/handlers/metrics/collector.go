package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// deliveryMetricKey name of the delivery per subscription metric
	deliveryMetricKey = "nats_ec_delivery_per_subscription_total"
	// eventTypeSubscribedMetricKey name of the eventType subscribed metric
	eventTypeSubscribedMetricKey = "nats_ec_event_type_subscribed_total"
	// deliveryMetricHelp help text for the delivery per subscription metric
	deliveryMetricHelp = "The total number of dispatched events per subscription"
	// eventTypeSubscribedMetricHelp help text for the eventType subscribed metric
	eventTypeSubscribedMetricHelp = "The total number of all the eventTypes subscribed using the Subscription CRD"
)

// Collector implements the prometheus.Collector interface
type Collector struct {
	deliveryPerSubscription *prometheus.CounterVec
	eventTypes              *prometheus.CounterVec
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
		eventTypes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: eventTypeSubscribedMetricKey,
				Help: eventTypeSubscribedMetricHelp,
			},
			[]string{"subscription_name", "subscription_namespace", "event_type", "consumer_name"},
		),
	}
}

// Describe implements the prometheus.Collector interface Describe method
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.deliveryPerSubscription.Describe(ch)
	c.eventTypes.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.deliveryPerSubscription.Collect(ch)
	c.eventTypes.Collect(ch)
}

// RegisterMetrics registers the metrics
func (c *Collector) RegisterMetrics() {
	metrics.Registry.MustRegister(c.deliveryPerSubscription)
	metrics.Registry.MustRegister(c.eventTypes)
}

// RecordDeliveryPerSubscription records a nats_ec_delivery_per_subscription_total metric
func (c *Collector) RecordDeliveryPerSubscription(subscriptionName, eventType, sink string, statusCode int) {
	c.deliveryPerSubscription.WithLabelValues(subscriptionName, eventType, fmt.Sprintf("%v", sink), fmt.Sprintf("%v", statusCode)).Inc()
}

// RecordEventTypes records a nats_ec_event_type_subscribed_total metric
func (c *Collector) RecordEventTypes(subscriptionName, subscriptionNamespace, eventType, consumer string) {
	c.eventTypes.WithLabelValues(subscriptionName, subscriptionNamespace, eventType, consumer).Inc()
}
