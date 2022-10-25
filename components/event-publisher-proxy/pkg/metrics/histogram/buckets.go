package histogram

//go:generate mockery --name BucketsProvider
type BucketsProvider interface {
	Buckets() []float64
}
