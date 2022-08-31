package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	FluentBitTriggeredRestartsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_fluentbit_triggered_restarts_total",
		Help: "Number of triggered Fluent Bit restarts",
	})
)

var registerMetrics sync.Once

func RegisterMetrics() {
	registerMetrics.Do(func() {
		metrics.Registry.MustRegister(FluentBitTriggeredRestartsTotal)
	})
}
