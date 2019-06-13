package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

//Unique Names of controller type for which statistics are collected
const (
	SbuController = "service_binding_usage_controller"
	UkController  = "usage_kind_controller"
)

//ControllerBusinessMetric represents metrics exporter
type ControllerBusinessMetric struct {
	errors  *prometheus.CounterVec
	queue   *prometheus.GaugeVec
	latency *prometheus.HistogramVec
}

//NewControllerBusinessMetric returns new ControllerBusinessMetric
func NewControllerBusinessMetric() *ControllerBusinessMetric {
	return &ControllerBusinessMetric{
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "controller_runtime_reconcile_errors_total",
			Help: "Total number of reconcilation errors per controller",
		}, []string{"controller"}),
		queue: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "controller_runtime_reconcile_queue_length",
			Help: "Length of reconcile queue per controller",
		}, []string{"controller"}),
		latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "controller_runtime_reconcile_latency",
			Help:    "Time latency of the duration of reconciliations per controller",
			Buckets: prometheus.LinearBuckets(0, 0.035, 15),
		}, []string{"controller"}),
	}
}

//Describe implementation of prometheus.Collector interface method
func (cbm *ControllerBusinessMetric) Describe(ch chan<- *prometheus.Desc) {
	cbm.errors.Describe(ch)
	cbm.queue.Describe(ch)
	cbm.latency.Describe(ch)
}

//Collect implementation of prometheus.Collector interface method
func (cbm *ControllerBusinessMetric) Collect(ch chan<- prometheus.Metric) {
	cbm.errors.Collect(ch)
	cbm.queue.Collect(ch)
	cbm.latency.Collect(ch)
}

//RecordError counts all errors during reconcile process
func (cbm *ControllerBusinessMetric) RecordError(controller string) {
	cbm.errors.WithLabelValues(controller).Inc()
}

//IncrementQueueLength increments gauge to estimate queue length
func (cbm *ControllerBusinessMetric) IncrementQueueLength(controller string) {
	cbm.queue.WithLabelValues(controller).Inc()
}

//DecrementQueueLength decrements gauge to estimate queue length
func (cbm *ControllerBusinessMetric) DecrementQueueLength(controller string) {
	cbm.queue.WithLabelValues(controller).Dec()
}

//RecordLatency saves probe to estimate reconcile process latency
func (cbm *ControllerBusinessMetric) RecordLatency(controller string, reconcileTime time.Duration) {
	cbm.latency.WithLabelValues(controller).Observe(reconcileTime.Seconds())
}
