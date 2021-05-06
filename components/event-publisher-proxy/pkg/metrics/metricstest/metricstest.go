// Package metricstest provides utilities for metrics testing.
package metricstest

import (
	"testing"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// EnsureMetricErrors ensures metric errors exists
func EnsureMetricErrors(t *testing.T, collector *metrics.Collector) {
	ensureMetricCount(t, collector, metrics.Errors, 1)
}

// EnsureMetricLatency ensures metric latency exists
func EnsureMetricLatency(t *testing.T, collector *metrics.Collector) {
	ensureMetricCount(t, collector, metrics.Latency, 1)
}

func ensureMetricCount(t *testing.T, collector *metrics.Collector, metric string, expectedCount int) {
	if count := testutil.CollectAndCount(collector, metric); count != expectedCount {
		t.Fatalf("invalid count for metric:%s, want:%d, got:%d", metric, expectedCount, count)
	}
}
