package summary

import "strings"

// By ...
type By func(r1, r2 *SpecificTestStats) bool

// ResultsSorter ...
type ResultsSorter struct {
	list []SpecificTestStats
	by   By
}

// NewResultsSorter ...
func NewResultsSorter(list []SpecificTestStats, by By) *ResultsSorter {
	return &ResultsSorter{
		list: list,
		by:   by,
	}
}

// Len is part of sort.Interface.
func (s *ResultsSorter) Len() int {
	return len(s.list)
}

// Swap is part of sort.Interface.
func (s *ResultsSorter) Swap(i, j int) {
	s.list[i], s.list[j] = s.list[j], s.list[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *ResultsSorter) Less(i, j int) bool {
	return s.by(&s.list[i], &s.list[j])
}

// ByName sorts by name
func ByName(r1, r2 *SpecificTestStats) bool {
	return strings.Compare(r1.Name, r2.Name) < 0
}

// ByMostFailures sorts by number of failures (most frequent first)
func ByMostFailures(r1, r2 *SpecificTestStats) bool {
	return r1.Failures < r2.Failures
}
