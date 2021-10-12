package collector

type Collector interface {
	AddObservation(observation float64, labelValues ...string)
}
