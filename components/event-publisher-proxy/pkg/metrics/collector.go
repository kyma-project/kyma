package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/histogram"
)

const (
	// HealthKey name of the health metric
	HealthKey = "eventing_epp_health"
	// HealthHelp help text for the Health metric.
	healthHelp = "The current health of the system. `1` indicates a healthy system."

	// BackendLatencyKey name of the backendLatency metric.
	BackendLatencyKey = "eventing_epp_backend_duration_milliseconds"
	// backendLatencyHelp help text for the backendLatency metric.
	backendLatencyHelp = "The duration of sending events to the messaging server in milliseconds"

	// durationKey name of the duration metric.
	durationKey = "eventing_epp_requests_duration_milliseconds"
	// durationHelp help text for the duration metric.
	durationHelp = "The duration of processing an incoming request (includes sending to the backend)"

	// RequestsKey name of the Requests metric.
	RequestsKey = "eventing_epp_requests_total"
	// requestsHelp help text for event requests metric.
	requestsHelp = "The total number of requests"

	// EventTypePublishedMetricKey name of the eventTypeLabel metric.
	EventTypePublishedMetricKey = "eventing_epp_event_type_published_total"
	// eventTypePublishedMetricHelp help text for the eventTypeLabel metric.
	eventTypePublishedMetricHelp = "The total number of events published for a given eventTypeLabel"
	// methodLabel label for the method used in the http request.

	methodLabel = "method"
	// responseCodeLabel name of the status code labels used by multiple metrics.
	responseCodeLabel = "code"
	// pathLabel name of the path service label.
	pathLabel = "path"
	// destSvcLabel name of the destination service label used by multiple metrics.
	destSvcLabel = "destination_service"
	// eventTypeLabel name of the event type label used by metrics.
	eventTypeLabel = "event_type"
	// eventSourceLabel name of the event source label used by metrics.
	eventSourceLabel = "event_source"
)

// PublishingMetricsCollector interface provides a Prometheus compatible Collector with additional convenience methods
// for recording epp specific metrics.
type PublishingMetricsCollector interface {
	prometheus.Collector
	RecordBackendLatency(duration time.Duration, statusCode int, destSvc string)
	RecordEventType(eventType, eventSource string, statusCode int)
	MetricsMiddleware() mux.MiddlewareFunc
}

var _ PublishingMetricsCollector = &Collector{}

// Collector implements the prometheus.Collector interface.
type Collector struct {
	backendLatency *prometheus.HistogramVec

	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec

	eventType *prometheus.CounterVec

	health *prometheus.GaugeVec
}

// NewCollector creates a new instance of Collector.
func NewCollector(latency histogram.BucketsProvider) *Collector {
	return &Collector{
		//nolint:promlinter // we follow the same pattern as istio. so a millisecond unit if fine here
		backendLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    BackendLatencyKey,
				Help:    backendLatencyHelp,
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

		//nolint:promlinter // we follow the same pattern as istio. so a millisecond unit if fine here
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    durationKey,
				Help:    durationHelp,
				Buckets: latency.Buckets(),
			},
			[]string{responseCodeLabel, methodLabel, pathLabel},
		),
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: RequestsKey,
				Help: requestsHelp,
			},
			[]string{responseCodeLabel, methodLabel, pathLabel},
		),
		health: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: HealthKey,
				Help: healthHelp,
			},
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface Describe method.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.backendLatency.Describe(ch)
	c.eventType.Describe(ch)
	c.requests.Describe(ch)
	c.duration.Describe(ch)
	c.health.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.backendLatency.Collect(ch)
	c.eventType.Collect(ch)
	c.requests.Collect(ch)
	c.duration.Collect(ch)
	c.health.Collect(ch)
}

// RecordLatency records a backendLatencyHelp metric.
func (c *Collector) RecordBackendLatency(duration time.Duration, statusCode int, destSvc string) {
	c.backendLatency.WithLabelValues(fmt.Sprint(statusCode), destSvc).Observe(float64(duration.Milliseconds()))
}

// RecordEventType updates the health status metric
func (c *Collector) SetHealthStatus(healthy bool) {
	var v float64
	if healthy {
		v = 1
	}
	c.health.WithLabelValues().Set(v)
}

// RecordEventType records an eventType metric.
func (c *Collector) RecordEventType(eventType, eventSource string, statusCode int) {
	c.eventType.WithLabelValues(eventType, eventSource, fmt.Sprint(statusCode)).Inc()
}

// MetricsMiddleware returns a http.Handler that can be used as middleware in gorilla.mux to track
// latencies for all handled paths in the gorilla router.
func (c *Collector) MetricsMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()
			promhttp.InstrumentHandlerDuration(
				c.duration.MustCurryWith(prometheus.Labels{pathLabel: path}),
				promhttp.InstrumentHandlerCounter(c.requests.MustCurryWith(prometheus.Labels{pathLabel: path}), next),
			).ServeHTTP(w, r)
		})
	}
}
