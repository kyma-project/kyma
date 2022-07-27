package metrics

import (
	"fmt"
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
	//EventRequests //todo
	EventRequests = "event_requests" //todo
	// errorsHelp help for the errors metric
	errorsHelp = "The total number of errors while sending Events to the messaging server"
	// latencyHelp help for the latency metric
	latencyHelp = "The duration of sending Events to the messaging server"
	// EventTypePublishedMetricHelp help for the eventType metric
	EventTypePublishedMetricHelp = "The total number of events published for a given eventType"
	//todo
	EventRequestsHelp = "lol" //todo
	//todo
	statusCode = "status_code"
	//todo
	destSvc = "destination_service"
)

// Collector implements the prometheus.Collector interface
type Collector struct {
	errors    *prometheus.CounterVec
	latency   *prometheus.HistogramVec
	eventType *prometheus.CounterVec
	requests  *prometheus.CounterVec
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
			[]string{statusCode, destSvc},
		),
		eventType: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: EventTypePublishedMetricKey,
				Help: EventTypePublishedMetricHelp,
			},
			[]string{"event_type", "event_source"},
		),
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: EventRequests,
				Help: EventTypePublishedMetricHelp,
			},
			[]string{statusCode, destSvc},
		),
	}
}

// Describe implements the prometheus.Collector interface Describe method
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.errors.Describe(ch)
	c.latency.Describe(ch)
	c.eventType.Describe(ch)
	c.requests.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.errors.Collect(ch)
	c.latency.Collect(ch)
	c.eventType.Collect(ch)
	c.requests.Collect(ch)
}

// RecordError records an error metric
func (c *Collector) RecordError() {
	c.errors.WithLabelValues().Inc()
}

// RecordLatency records a latency metric
func (c *Collector) RecordLatency(duration time.Duration, statusCode int, destSvc string) {
	c.latency.WithLabelValues(fmt.Sprint(statusCode), destSvc).Observe(float64(duration.Milliseconds()))
}

// RecordEventType records a eventType metric
func (c *Collector) RecordEventType(eventType, eventSource string) {
	c.eventType.WithLabelValues(eventType, eventSource).Inc()
}

// todo
func (c *Collector) RecordRequests(statusCode int, destSvc string) {
	c.requests.WithLabelValues(fmt.Sprint(statusCode), destSvc).Inc()
}
