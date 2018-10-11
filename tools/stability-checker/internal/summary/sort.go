package summary


type ByMostFailures []SpecificTestStats

// Len is part of sort.Interface.
func (s ByMostFailures) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s ByMostFailures) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s ByMostFailures) Less(i, j int) bool {
	return s[i].Failures > s[j].Failures
}

