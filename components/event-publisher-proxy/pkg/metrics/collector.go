package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/histogram"
)

const (
	// ErrorsKey name of the errors metric
	ErrorsKey = "eventing_epp_errors_total"
	// LatencyKey name of the latency metric
	LatencyKey = "eventing_epp_messaging_server_latency_duration_milliseconds"
	// EventTypePublishedMetricKey name of the eventTypeLabel metric
	EventTypePublishedMetricKey = "nats_epp_event_type_published_total"
	//EventRequestsKey name of the eventRequests metric
	EventRequestsKey = "eventing_epp_requests_total"
	// errorsHelp help text for the errors metric
	errorsHelp = "The total number of errors while sending events to the messaging server"
	// latencyHelp help text for the latency metric
	latencyHelp = "The duration of sending events to the messaging server in milliseconds"
	// eventTypePublishedMetricHelp help text for the eventTypeLabel metric
	eventTypePublishedMetricHelp = "The total number of events published for a given eventTypeLabel"
	// eventRequestsHelp help text for event requests metric
	eventRequestsHelp = "The total number of event requests"
	//responseCodeLabel name of the status code labels used by multiple metrics
	responseCodeLabel = "response_code"
	//destSvcLabel name of the destination service label used by multiple metrics
	destSvcLabel = "destination_service"
	// eventTypeLabel name of the event type label used by metrics
	eventTypeLabel = "event_type"
	// eventSourceLabel name of the event source label used by metrics
	eventSourceLabel = "event_source"
)

// Collector implements the prometheus.Collector interface
type Collector struct {
	errors    *prometheus.CounterVec
	latency   *prometheus.HistogramVec
	eventType *prometheus.CounterVec
	requests  *prometheus.CounterVec
}

// NewCollector a new instance of Collector
func NewCollector(latency histogram.BucketsProvider) *Collector {
	return &Collector{
		errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: ErrorsKey,
				Help: errorsHelp,
			},
			[]string{},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    LatencyKey,
				Help:    latencyHelp,
				Buckets: latency.Buckets(),
			},
			[]string{responseCodeLabel, destSvcLabel},
		),
		eventType: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: EventTypePublishedMetricKey,
				Help: eventTypePublishedMetricHelp,
			},
			[]string{eventTypeLabel, eventSourceLabel, responseCodeLabel},
		),
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: EventRequestsKey,
				Help: eventRequestsHelp,
			},
			[]string{responseCodeLabel, destSvcLabel},
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

// RecordEventType records an eventType metric
func (c *Collector) RecordEventType(eventType, eventSource string, statusCode int) {
	c.eventType.WithLabelValues(eventType, eventSource, fmt.Sprint(statusCode)).Inc()
}

// RecordRequests records an eventRequests metric
func (c *Collector) RecordRequests(statusCode int, destSvc string) {
	c.requests.WithLabelValues(fmt.Sprint(statusCode), destSvc).Inc()
}
