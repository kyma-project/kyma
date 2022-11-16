// Package metricstest provides utilities for metrics testing.
package metricstest

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
)

// EnsureMetricErrors ensures metric eventing_epp_errors_total exists.
func EnsureMetricErrors(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.ErrorsKey, count)
}

// EnsureMetricLatency ensures metric eventing_epp_messaging_server_latency_duration_milliseconds exists.
func EnsureMetricLatency(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.LatencyKey, count)
}

// EnsureMetricEventTypePublished ensures metric epp_event_type_published_total exists.
func EnsureMetricEventTypePublished(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.EventTypePublishedMetricKey, count)
}

// EnsureMetricTotalRequests ensures metric eventing_epp_requests_total exists.
func EnsureMetricTotalRequests(t *testing.T, collector metrics.PublishingMetricsCollector, count int) {
	ensureMetricCount(t, collector, metrics.EventRequestsKey, count)
}

func ensureMetricCount(t *testing.T, collector metrics.PublishingMetricsCollector, metric string, expectedCount int) {
	if count := testutil.CollectAndCount(collector, metric); count != expectedCount {
		t.Fatalf("invalid count for metric:%s, want:%d, got:%d", metric, expectedCount, count)
	}
}

// EnsureMetricMatchesTextExpositionFormat ensures that metrics collected by the given collector match the given metric output in TextExpositionFormat.
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
