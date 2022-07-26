package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Errors name of the errors metric
	Errors = "event_publish_to_messaging_server_errors_total"
	// Latency name of the latency metric
	Latency = "event_publish_to_messaging_server_latency"
	// EventTypePublishedMetricKey name of the eventType metric
	EventTypePublishedMetricKey = "event_type_published"
	// errorsHelp help for the errors metric
	errorsHelp = "The total number of errors while sending Events to the messaging server"
	// latencyHelp help for the latency metric
	latencyHelp = "The duration of sending Events to the messaging server"
	// EventTypePublishedMetricHelp help for the eventType metric
	EventTypePublishedMetricHelp = "The total number of events published for a given eventType"
)

// Collector implements the prometheus.Collector interface
type Collector struct {
	errors    *prometheus.CounterVec
	latency   *prometheus.HistogramVec
	eventType *prometheus.CounterVec
}

// NewCollector a new instance of Collector
func NewCollector() *Collector {
	return &Collector{
		errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: Errors,
				Help: errorsHelp,
			},
			[]string{},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: Latency,
				Help: latencyHelp,
			},
			[]string{"status_code", "destination_service"},
		),
		eventType: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: EventTypePublishedMetricKey,
				Help: EventTypePublishedMetricHelp,
			},
			[]string{"event_type", "event_source"},
		),
	}
}

// Describe implements the prometheus.Collector interface Describe method
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.errors.Describe(ch)
	c.latency.Describe(ch)
	c.eventType.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.errors.Collect(ch)
	c.latency.Collect(ch)
	c.eventType.Collect(ch)
}

// RecordError records an error metric
func (c *Collector) RecordError() {
	c.errors.WithLabelValues().Inc()
}

// RecordLatency records a latency metric
func (c *Collector) RecordLatency(duration time.Duration, statusCode int, destinationService string) {
	c.latency.WithLabelValues(string(statusCode), destinationService).Observe(float64(duration.Microseconds()))
}

// RecordEventType records a eventType metric
func (c *Collector) RecordEventType(eventType, eventSource string) {
	c.eventType.WithLabelValues(eventType, eventSource).Inc()
}
