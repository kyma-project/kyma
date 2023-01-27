package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// latencyMetricKey name of the dispatch_duration metric
	latencyMetricKey = "eventing_ec_nats_subscriber_dispatch_duration_seconds"
	// latencyMetricHelp help text for the dispatch_duration metric
	latencyMetricHelp = "The duration of sending an incoming nats message to the subscriber (not including processing the message in the dispatcher)"
	// deliveryMetricKey name of the delivery per subscription metric.
	deliveryMetricKey = "eventing_ec_nats_delivery_per_subscription_total"
	// eventTypeSubscribedMetricKey name of the eventType subscribed metric.
	eventTypeSubscribedMetricKey = "eventing_ec_event_type_subscribed_total"
	// deliveryMetricHelp help text for the delivery per subscription metric.
	deliveryMetricHelp = "The total number of dispatched events per subscription"
	// eventTypeSubscribedMetricHelp help text for the eventType subscribed metric.
	eventTypeSubscribedMetricHelp = "The total number of eventTypes subscribed using the Subscription CRD"
)

// Collector implements the prometheus.Collector interface.
type Collector struct {
	deliveryPerSubscription *prometheus.CounterVec
	eventTypes              *prometheus.CounterVec
	latencyPerSubscriber    *prometheus.HistogramVec
}

// NewCollector a new instance of Collector.
func NewCollector() *Collector {
	return &Collector{
		deliveryPerSubscription: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: deliveryMetricKey,
				Help: deliveryMetricHelp,
			},
			[]string{"subscription_name", "event_type", "sink", "response_code"},
		),
		latencyPerSubscriber: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    latencyMetricKey,
				Help:    latencyMetricHelp,
				Buckets: prometheus.ExponentialBuckets(0.002, 2, 10),
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

// Describe implements the prometheus.Collector interface Describe method.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.deliveryPerSubscription.Describe(ch)
	c.eventTypes.Describe(ch)
	c.latencyPerSubscriber.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.deliveryPerSubscription.Collect(ch)
	c.eventTypes.Collect(ch)
	c.latencyPerSubscriber.Collect(ch)
}

// RegisterMetrics registers the metrics
func (c *Collector) RegisterMetrics() {
	metrics.Registry.MustRegister(c.deliveryPerSubscription)
	metrics.Registry.MustRegister(c.eventTypes)
	metrics.Registry.MustRegister(c.latencyPerSubscriber)
}

// RecordDeliveryPerSubscription records a eventing_ec_nats_delivery_per_subscription_total metric.
func (c *Collector) RecordDeliveryPerSubscription(subscriptionName, eventType, sink string, statusCode int) {
	c.deliveryPerSubscription.WithLabelValues(subscriptionName, eventType, fmt.Sprintf("%v", sink), fmt.Sprintf("%v", statusCode)).Inc()
}

// RecordLatencyPerSubscription records a eventing_ec_nats_subscriber_dispatch_duration_seconds
func (c *Collector) RecordLatencyPerSubscription(duration time.Duration, subscriptionName, eventType, sink string, statusCode int) {
	c.latencyPerSubscriber.WithLabelValues(subscriptionName, eventType, fmt.Sprintf("%v", sink), fmt.Sprintf("%v", statusCode)).Observe(duration.Seconds())
}

// RecordEventTypes records a eventing_ec_event_type_subscribed_total metric.
func (c *Collector) RecordEventTypes(subscriptionName, subscriptionNamespace, eventType, consumer string) {
	c.eventTypes.WithLabelValues(subscriptionName, subscriptionNamespace, eventType, consumer).Inc()
}
