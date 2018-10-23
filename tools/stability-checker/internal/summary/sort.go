package summary

// ByMostFailures provides order of SpecificTestStats where most failed tests are placed at the begging of the slice
type ByMostFailures []SpecificTestStats

// Len is part of sort.Interface.
func (s ByMostFailures) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s ByMostFailures) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface.
func (s ByMostFailures) Less(i, j int) bool {
	return s[i].Failures > s[j].Failures
}
