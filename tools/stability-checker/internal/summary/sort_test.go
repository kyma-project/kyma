package summary_test

import (
	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestSortByNames(t *testing.T) {
	// GIVEN
	sut := summary.NewResultsSorter(fixTestStatsList(), summary.ByName)
	// WHEN
	sort.Sort(sut)
	// THEN
	assert.Len(t,sut.List, 3)
	assert.Equal(t,sut.List[0].Name, "a")
	assert.Equal(t,sut.List[1].Name, "d")
	assert.Equal(t, sut.List[2].Name, "z")

}

func TestSortByMostFailures(t *testing.T) {
	// GIVEN
	sut := summary.NewResultsSorter(fixTestStatsList(), summary.ByMostFailures)
	// WHEN
	sort.Sort(sut)
	// THEN
	assert.Len(t,sut.List, 3)
	assert.Equal(t,sut.List[0].Failures, 10)
	assert.Equal(t,sut.List[1].Failures, 7)
	assert.Equal(t, sut.List[2].Failures, 3)
}

func fixTestStatsList() []summary.SpecificTestStats {
	return []summary.SpecificTestStats{
		{
			Name:     "d",
			Failures: 10,
		},
		{
			Name:     "a",
			Failures: 3,
		},
		{
			Name:     "z",
			Failures: 7,
		},
	}
}
