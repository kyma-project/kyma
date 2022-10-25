package latency

import "github.com/prometheus/client_golang/prometheus"

const (
	// Note: The following configuration generates the histogram buckets [2 4 8 16 32 64 128 256 512 1024].

	// start value of the prometheus.ExponentialBuckets start parameter.
	start float64 = 2.0
	// factor value of the prometheus.ExponentialBuckets factor parameter.
	factor float64 = 2.0
	// count value of the prometheus.ExponentialBuckets count parameter.
	count int = 10
)

type BucketsProvider struct {
	buckets []float64
}

func NewBucketsProvider() *BucketsProvider {
	return &BucketsProvider{buckets: prometheus.ExponentialBuckets(start, factor, count)}
}

func (p BucketsProvider) Buckets() []float64 {
	return p.buckets
}
