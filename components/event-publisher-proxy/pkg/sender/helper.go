package sender

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PublishingMetricsCollectorStub struct {
}

func (p PublishingMetricsCollectorStub) Describe(chan<- *prometheus.Desc) {
}

func (p PublishingMetricsCollectorStub) Collect(chan<- prometheus.Metric) {
}

func (p PublishingMetricsCollectorStub) RecordError() {
}

func (p PublishingMetricsCollectorStub) RecordLatency(time.Duration, int, string) {
}

func (p PublishingMetricsCollectorStub) RecordEventType(string, string, int) {
}

func (p PublishingMetricsCollectorStub) RecordRequests(int, string) {
}
