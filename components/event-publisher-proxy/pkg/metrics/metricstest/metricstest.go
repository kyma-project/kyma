// Package metricstest provides utilities for metrics testing.
package metricstest

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
)

// EnsureMetricErrors ensures metric eventing_epp_backend_errors_total exists.
func EnsureMetricErrors(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.BackendErrorsKey, count)
}

// EnsureMetricLatency ensures metric eventing_epp_backend_duration_seconds exists.
func EnsureMetricLatency(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.BackendLatencyKey, count)
}

// EnsureMetricEventTypePublished ensures metric eventing_epp_event_type_published_total exists.
func EnsureMetricEventTypePublished(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.EventTypePublishedMetricKey, count)
}

// EnsureMetricTotalRequests ensures metric eventing_epp_backend_requests_total exists.
func EnsureMetricTotalRequests(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.BackendRequestsKey, count)
}

func ensureMetricCount(t *testing.T, collector metrics.PublishingMetricsCollector, metric string, expectedCount int) {
	if count := testutil.CollectAndCount(collector, metric); count != expectedCount {
		t.Fatalf("invalid count for metric:%s, want:%d, got:%d", metric, expectedCount, count)
	}
}

// EnsureMetricMatchesTextExpositionFormat ensures that metrics collected by the given collector
// match the given metric output in TextExpositionFormat.
// This is useful to compare metrics with their given labels.
func EnsureMetricMatchesTextExpositionFormat(t *testing.T, collector metrics.PublishingMetricsCollector, tef string, metricNames ...string) {
	if err := testutil.CollectAndCompare(collector, strings.NewReader(tef), metricNames...); err != nil {
		t.Fatalf("%v", err)
	}
}

type PublishingMetricsCollectorStub struct {
}

func (p PublishingMetricsCollectorStub) Describe(_ chan<- *prometheus.Desc) {
}

func (p PublishingMetricsCollectorStub) Collect(_ chan<- prometheus.Metric) {
}

func (p PublishingMetricsCollectorStub) RecordError() {
}

func (p PublishingMetricsCollectorStub) RecordLatency(_ time.Duration, _ int, _ string) {
}

func (p PublishingMetricsCollectorStub) RecordEventType(_, _ string, _ int) {
}

func (p PublishingMetricsCollectorStub) RecordRequests(_ int, _ string) {
}

//nolint:lll // that's how TEF has to look like
func MakeTEFBackendDuration(code int, service string) string {
	tef := strings.ReplaceAll(`# HELP eventing_epp_backend_duration_milliseconds The duration of sending events to the messaging server in milliseconds
					# TYPE eventing_epp_backend_duration_milliseconds histogram
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="0.005"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="0.01"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="0.025"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="0.05"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="0.1"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="0.25"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="0.5"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="1"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="2.5"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="5"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="10"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="%%code%%",destination_service="%%service%%",le="+Inf"} 1
					eventing_epp_backend_duration_milliseconds_sum{code="%%code%%",destination_service="%%service%%"} 0
					eventing_epp_backend_duration_milliseconds_count{code="%%code%%",destination_service="%%service%%"} 1
					`, "%%code%%", strconv.Itoa(code))
	return strings.ReplaceAll(tef, "%%service%%", service)
}

func MakeTEFBackendRequests(code int, service string) string {
	tef := strings.ReplaceAll(`# HELP eventing_epp_backend_requests_total The total number of backend requests
					# TYPE eventing_epp_backend_requests_total counter
					eventing_epp_backend_requests_total{code="%%code%%",destination_service="%%service%%"} 1
					`, "%%code%%", strconv.Itoa(code))
	return strings.ReplaceAll(tef, "%%service%%", service)
}

//nolint:lll // that's how TEF has to look like
func MakeTEFBackendErrors() string {
	return `# HELP eventing_epp_backend_errors_total The total number of backend errors while sending events to the messaging server
        # TYPE eventing_epp_backend_errors_total counter
        eventing_epp_backend_errors_total 1
		`
}

//nolint:lll // that's how TEF has to look like
func MakeTEFEventTypePublished(code int, source, eventtype string) string {
	tef := strings.ReplaceAll(`# HELP eventing_epp_event_type_published_total The total number of events published for a given eventTypeLabel
        # TYPE eventing_epp_event_type_published_total counter
        eventing_epp_event_type_published_total{code="204",event_source="%%source%%",event_type="%%type%%"} 1
					`, "%%code%%", strconv.Itoa(code))
	tef = strings.ReplaceAll(tef, "%%source%%", source)
	return strings.ReplaceAll(tef, "%%type%%", eventtype)
}
