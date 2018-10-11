package summary_test

import (
	"sort"
	"testing"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/stretchr/testify/assert"
)



func TestSortByMostFailures(t *testing.T) {
	// GIVEN
	sut := summary.ByMostFailures(fixTestStatsList())
	// WHEN
	sort.Sort(sut)
	// THEN
	assert.Len(t, sut, 3)
	assert.Equal(t, sut[0].Failures, 10)
	assert.Equal(t, sut[1].Failures, 7)
	assert.Equal(t, sut[2].Failures, 3)
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
