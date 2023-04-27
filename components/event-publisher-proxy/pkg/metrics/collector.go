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
	// BackendErrorsKey name of the backendErrors metric.
	BackendErrorsKey = "eventing_epp_backend_errors_total"
	// backendErrorsHelp help text for the backendErrors metric.
	backendErrorsHelp = "The total number of backend errors while sending events to the messaging server"

	// BackendLatencyKey name of the backendLatencyHelp metric.
	BackendLatencyKey = "eventing_epp_backend_duration_milliseconds"
	// backendLatencyHelp help text for the backendLatencyHelp metric.
	backendLatencyHelp = "The duration of sending events to the messaging server in milliseconds"

	// BackendRequestsKey name of the eventRequests metric.
	BackendRequestsKey = "eventing_epp_backend_requests_total"
	// backendRequestsHelp help text for event backendRequests metric.
	backendRequestsHelp = "The total number of backend requests"

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
	RecordBackendError()
	RecordBackendLatency(duration time.Duration, statusCode int, destSvc string)
	RecordEventType(eventType, eventSource string, statusCode int)
	RecordBackendRequests(statusCode int, destSvc string)
	MetricsMiddleware() mux.MiddlewareFunc
}

var _ PublishingMetricsCollector = &Collector{}

// Collector implements the prometheus.Collector interface.
type Collector struct {
	backendErrors   *prometheus.CounterVec
	backendLatency  *prometheus.HistogramVec
	backendRequests *prometheus.CounterVec

	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec

	eventType *prometheus.CounterVec
}

// NewCollector creates a new instance of Collector.
func NewCollector(latency histogram.BucketsProvider) *Collector {
	return &Collector{
		backendErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: BackendErrorsKey,
				Help: backendErrorsHelp,
			},
			[]string{},
		),
		//nolint:promlinter // we follow the same pattern as istio. so a millisecond unit if fine here
		backendLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    BackendLatencyKey,
				Help:    backendLatencyHelp,
				Buckets: latency.Buckets(),
			},
			[]string{responseCodeLabel, destSvcLabel},
		),
		backendRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: BackendRequestsKey,
				Help: backendRequestsHelp,
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
			[]string{responseCodeLabel, methodLabel, pathLabel}),
	}
}

// Describe implements the prometheus.Collector interface Describe method.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.backendErrors.Describe(ch)
	c.backendLatency.Describe(ch)
	c.backendRequests.Describe(ch)
	c.eventType.Describe(ch)
	c.requests.Describe(ch)
	c.duration.Describe(ch)
}

// Collect implements the prometheus.Collector interface Collect method.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.backendErrors.Collect(ch)
	c.backendLatency.Collect(ch)
	c.backendRequests.Collect(ch)
	c.eventType.Collect(ch)
	c.requests.Collect(ch)
	c.duration.Collect(ch)
}

// RecordBackendError records an error while sending to the eventing backend.
func (c *Collector) RecordBackendError() {
	c.backendErrors.WithLabelValues().Inc()
}

// RecordLatency records a backendLatencyHelp metric.
func (c *Collector) RecordBackendLatency(duration time.Duration, statusCode int, destSvc string) {
	c.backendLatency.WithLabelValues(fmt.Sprint(statusCode), destSvc).Observe(float64(duration.Milliseconds()))
}

// RecordEventType records an eventType metric.
func (c *Collector) RecordEventType(eventType, eventSource string, statusCode int) {
	c.eventType.WithLabelValues(eventType, eventSource, fmt.Sprint(statusCode)).Inc()
}

// RecordRequests records an eventRequests metric.
func (c *Collector) RecordBackendRequests(statusCode int, destSvc string) {
	c.backendRequests.WithLabelValues(fmt.Sprint(statusCode), destSvc).Inc()
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
