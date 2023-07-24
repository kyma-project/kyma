package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// healthMetricKey name of the health metric.
	healthMetricKey = "eventing_ec_health"
	// healthMetricHelp help text for the Health metric.
	healthMetricHelp = "The current health of the system. `1` indicates a healthy system"

	// latencyMetricKey name of the dispatch_duration metric.
	latencyMetricKey = "eventing_ec_nats_subscriber_dispatch_duration_seconds"
	//nolint:lll // help text for metrics
	// latencyMetricHelp help text for the dispatch_duration metric.
	latencyMetricHelp = "The duration of sending an incoming NATS message to the subscriber (not including processing the message in the dispatcher)"

	// deliveryMetricKey name of the delivery per subscription metric.
	deliveryMetricKey = "eventing_ec_nats_delivery_per_subscription_total"
	// deliveryMetricHelp help text for the delivery per subscription metric.
	deliveryMetricHelp = "The total number of dispatched events per subscription"

	// eventTypeSubscribedMetricKey name of the eventType subscribed metric.
	eventTypeSubscribedMetricKey = "eventing_ec_event_type_subscribed_total"
	// eventTypeSubscribedMetricHelp help text for the eventType subscribed metric.
	eventTypeSubscribedMetricHelp = "The total number of eventTypes subscribed using the Subscription CRD"

	// subscriptionStatus name of the subscription status metric.
	subscriptionStatusMetricKey = "eventing_ec_subscription_status"
	// subscriptionStatusMetricHelp help text for the subscription status metric.
	subscriptionStatusMetricHelp = "The status of a subscription. `1` indicates the subscription is marked as ready"

	subscriptionNameLabel      = "subscription_name"
	eventTypeLabel             = "event_type"
	sinkLabel                  = "sink"
	responseCodeLabel          = "response_code"
	subscriptionNamespaceLabel = "subscription_namespace"
	consumerNameLabel          = "consumer_name"
	backendTypeLabel           = "eventing_backend"
	streamNameLabel            = "stream_name"
)

// Collector implements the prometheus.Collector interface.
type Collector struct {
	deliveryPerSubscription *prometheus.CounterVec
	eventTypes              *prometheus.CounterVec
	latencyPerSubscriber    *prometheus.HistogramVec
	health                  *prometheus.GaugeVec
	subscriptionStatus      *prometheus.GaugeVec
}

// NewCollector a new instance of Collector.
func NewCollector() *Collector {
	return &Collector{
		deliveryPerSubscription: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: deliveryMetricKey,
				Help: deliveryMetricHelp,
			},
			[]string{subscriptionNameLabel, eventTypeLabel, sinkLabel, responseCodeLabel},
		),
		latencyPerSubscriber: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    latencyMetricKey,
				Help:    latencyMetricHelp,
				Buckets: prometheus.ExponentialBuckets(0.002, 2, 10),
			},
			[]string{subscriptionNameLabel, eventTypeLabel, sinkLabel, responseCodeLabel},
		),
		eventTypes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: eventTypeSubscribedMetricKey,
				Help: eventTypeSubscribedMetricHelp,
			},
			[]string{subscriptionNameLabel, subscriptionNamespaceLabel, eventTypeLabel, consumerNameLabel},
		),
		health: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: healthMetricKey,
				Help: healthMetricHelp,
			},
			nil,
		),
		subscriptionStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: subscriptionStatusMetricKey,
				Help: subscriptionStatusMetricHelp,
			},
			[]string{subscriptionNameLabel, subscriptionNamespaceLabel, consumerNameLabel, backendTypeLabel, streamNameLabel},
		),
	}
}

// Describe implements the prometheus.Collector interface Describe method.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.deliveryPerSubscription.Describe(ch)
	c.eventTypes.Describe(ch)
	c.latencyPerSubscriber.Describe(ch)
	c.health.Describe(ch)
	c.subscriptionStatus.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.deliveryPerSubscription.Collect(ch)
	c.eventTypes.Collect(ch)
	c.latencyPerSubscriber.Collect(ch)
	c.health.Collect(ch)
	c.subscriptionStatus.Collect(ch)
}

// RegisterMetrics registers the metrics.
func (c *Collector) RegisterMetrics() {
	metrics.Registry.MustRegister(c.deliveryPerSubscription)
	metrics.Registry.MustRegister(c.eventTypes)
	metrics.Registry.MustRegister(c.latencyPerSubscriber)
	metrics.Registry.MustRegister(c.health)
	metrics.Registry.MustRegister(c.subscriptionStatus)

	// set health metric to 1. With future updates this can be tied to other health indicators.
	c.health.WithLabelValues().Set(1)
}

// RecordDeliveryPerSubscription records a eventing_ec_nats_delivery_per_subscription_total metric.
func (c *Collector) RecordDeliveryPerSubscription(subscriptionName, eventType, sink string, statusCode int) {
	c.deliveryPerSubscription.WithLabelValues(
		subscriptionName,
		eventType,
		fmt.Sprintf("%v", sink),
		fmt.Sprintf("%v", statusCode)).Inc()
}

// RecordLatencyPerSubscription records a eventing_ec_nats_subscriber_dispatch_duration_seconds.
func (c *Collector) RecordLatencyPerSubscription(
	duration time.Duration,
	subscriptionName, eventType, sink string,
	statusCode int) {
	c.latencyPerSubscriber.WithLabelValues(
		subscriptionName,
		eventType,
		fmt.Sprintf("%v", sink),
		fmt.Sprintf("%v", statusCode)).Observe(duration.Seconds())
}

// RecordEventTypes records a eventing_ec_event_type_subscribed_total metric.
func (c *Collector) RecordEventTypes(subscriptionName, subscriptionNamespace, eventType, consumer string) {
	c.eventTypes.WithLabelValues(subscriptionName, subscriptionNamespace, eventType, consumer).Inc()
}

// RecordSubscriptionStatus records an eventing_ec_subscription_status metric.
func (c *Collector) RecordSubscriptionStatus(isActive bool, subscriptionName,
	subscriptionNamespace, backendType, consumer, streamName string) {
	var v float64
	if isActive {
		v = 1
	}
	c.subscriptionStatus.With(prometheus.Labels{
		subscriptionNameLabel:      subscriptionName,
		subscriptionNamespaceLabel: subscriptionNamespace,
		consumerNameLabel:          consumer,
		backendTypeLabel:           backendType,
		streamNameLabel:            streamName,
	}).Set(v)
}

// RemoveSubscriptionStatus removes an eventing_ec_subscription_status metric.
func (c *Collector) RemoveSubscriptionStatus(subscriptionName, subscriptionNamespace,
	backendType, consumer, streamName string) {
	c.subscriptionStatus.Delete(prometheus.Labels{
		subscriptionNameLabel:      subscriptionName,
		subscriptionNamespaceLabel: subscriptionNamespace,
		consumerNameLabel:          consumer,
		backendTypeLabel:           backendType,
		streamNameLabel:            streamName,
	})
}

func (c *Collector) ResetSubscriptionStatus() {
	c.subscriptionStatus.Reset()
}
