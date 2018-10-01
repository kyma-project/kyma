package summary

// SpecificTestStats aggregates information how many specific test ends up with success and failure.
type SpecificTestStats struct {
	Name      string
	Successes int
	Failures  int
}

// add adds test result
func (b *SpecificTestStats) add(val bool) {
	if val {
		b.Successes++
	} else {
		b.Failures++
	}
}

// statsAggregator aggregates test results
type statsAggregator struct {
	m map[string]*SpecificTestStats
}

func newStatsAggregator() *statsAggregator {
	return &statsAggregator{
		m: make(map[string]*SpecificTestStats),
	}
}

func (sc *statsAggregator) AddTestResult(testName string, result bool) {
	_, ex := sc.m[testName]
	if ex {
		sc.m[testName].add(result)
	} else {
		newTest := &SpecificTestStats{Name: testName}
		newTest.add(result)
		sc.m[testName] = newTest
	}
}

func (sc *statsAggregator) Merge(in map[string]SpecificTestStats) {
	for k, v := range in {
		curr, ex := sc.m[k]
		if ex {
			curr.Successes += v.Successes
			curr.Failures += v.Failures
		} else {
			sc.m[k] = &SpecificTestStats{Name: k, Failures: v.Failures, Successes: v.Successes}
		}
	}
}

func (sc *statsAggregator) ToMap() map[string]SpecificTestStats {
	out := make(map[string]SpecificTestStats)
	for k, v := range sc.m {
		out[k] = *v
	}
	return out
}

func (sc *statsAggregator) ToList() []SpecificTestStats {
	out := make([]SpecificTestStats, len(sc.m))
	i := 0
	for _, v := range sc.m {
		out[i] = *v
		i++
	}
	return out
}
