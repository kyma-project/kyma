package monitoring

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	reqCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all HTTP requests.",
	}, []string{"code", "method"})

	reqDurations = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "requests_durations",
			Help: "Requests latencies in seconds",
		})

	authnDurations = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "authentication_durations",
			Help: "Requests authentication latencies in seconds",
		})

	authzDurations = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "authorization_durations",
			Help: "Requests authorization latencies in seconds",
		})

	spdyNegotiationDurations = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "spdy_negotiation_durations",
			Help: "SPDY negotiation latencies in seconds",
		})
)

type ProxyMetrics struct {
	RequestCounterVec       *prometheus.CounterVec
	RequestDurations        prometheus.Summary
	AuthenticationDurations prometheus.Summary
	AuthorizationDurations  prometheus.Summary
}

//NewProxyMetrics returns registered Prometheus metrics for the proxy
func NewProxyMetrics() (*ProxyMetrics, error) {
	err := registerProxyMetrics()
	if err != nil {
		return nil, err
	}
	return &ProxyMetrics{
		RequestCounterVec:       reqCounter,
		RequestDurations:        reqDurations,
		AuthenticationDurations: authnDurations,
		AuthorizationDurations:  authzDurations,
	}, nil
}

func registerProxyMetrics() error {
	err := prometheus.Register(reqCounter)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not register metric: %s", err))
	}

	err = prometheus.Register(reqDurations)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not register metric: %s", err))
	}

	err = prometheus.Register(authnDurations)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not register metric: %s", err))
	}

	err = prometheus.Register(authzDurations)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not register metric: %s", err))
	}

	return nil
}

type SPDYMetrics struct {
	NegotiationDurations prometheus.Summary
}

//NewSPDYMetrics returns registered Prometheus metric for SPDY
func NewSPDYMetrics() *SPDYMetrics {
	registerSPDYMetrics()
	return &SPDYMetrics{NegotiationDurations: spdyNegotiationDurations}
}

func registerSPDYMetrics() {
	prometheus.MustRegister(spdyNegotiationDurations)
}
