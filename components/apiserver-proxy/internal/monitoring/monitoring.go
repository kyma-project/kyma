package monitoring

import "github.com/prometheus/client_golang/prometheus"

var (
	reqCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total number of requests."})

	reqCounterByCode = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total_code",
			Help: "Total number of requests, partitioned by status code.",
		},
		[]string{"code"})

	reqDurations = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "requests_durations",
			Help:       "Requests latencies in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})

	authnDurations = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "authentication_durations",
			Help:       "Requests authentication latencies in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})

	authzDurations = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "authorization_durations",
			Help:       "Requests authorization latencies in seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})
)

type ProxyMetrics struct {
	RequestCounter          prometheus.Counter
	RequestByCodeCounter    *prometheus.CounterVec
	RequestDurations        prometheus.Summary
	AuthenticationDurations prometheus.Summary
	AuthorizationDurations  prometheus.Summary
}

//NewProxyMetrics returns registered Prometheus metrics for the proxy
func NewProxyMetrics() *ProxyMetrics {
	registerMetrics()
	return &ProxyMetrics{
		RequestCounter:          reqCounter,
		RequestByCodeCounter:    reqCounterByCode,
		RequestDurations:        reqDurations,
		AuthenticationDurations: authnDurations,
		AuthorizationDurations:  authzDurations,
	}
}

func registerMetrics() {
	prometheus.MustRegister(reqCounter)
	prometheus.MustRegister(reqCounterByCode)
	prometheus.MustRegister(reqDurations)
	prometheus.MustRegister(authnDurations)
	prometheus.MustRegister(authzDurations)
}
