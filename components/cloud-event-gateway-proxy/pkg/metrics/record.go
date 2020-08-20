package metrics

// Record stores the given Measurement from `ms` in the current metrics backend.
//func Record(ctx context.Context, ms stats.Measurement, ros ...stats.Options) {
//	getCurMetricsConfig().record(ctx, []stats.Measurement{ms}, ros...)
//}

// RecordBatch stores the given Measurements from `mss` in the current metrics backend.
// All metrics should be reported using the same Resource.
//func RecordBatch(ctx context.Context, mss ...stats.Measurement) {
//	getCurMetricsConfig().record(ctx, mss)
//}

// Buckets125 generates an array of buckets with approximate powers-of-two
// buckets that also aligns with powers of 10 on every 3rd step. This can
// be used to create a view.Distribution.
func Buckets125(low, high float64) []float64 {
	buckets := []float64{low}
	for last := low; last < high; last *= 10 {
		buckets = append(buckets, 2*last, 5*last, 10*last)
	}
	return buckets
}

// BucketsNBy10 generates an array of N buckets starting from low and
// multiplying by 10 n times.
func BucketsNBy10(low float64, n int) []float64 {
	buckets := []float64{low}
	for last, i := low, len(buckets); i < n; last, i = 10*last, i+1 {
		buckets = append(buckets, 10*last)
	}
	return buckets
}
