package summary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecificTestStatsAdd(t *testing.T) {
	// GIVEN
	obj := SpecificTestStats{Name: "test"}
	// WHEN
	obj.add(true)
	obj.add(true)
	obj.add(false)
	// THEN
	assert.Equal(t, 2, obj.Successes)
	assert.Equal(t, 1, obj.Failures)
}

func TestStatsAggregatorAdd(t *testing.T) {
	// GIVEN
	sut := newStatsAggregator()
	// WHEN
	sut.AddTestResult("test-1", true)
	sut.AddTestResult("test-1", true)
	sut.AddTestResult("test-1", false)
	sut.AddTestResult("test-2", true)
	// THEN
	actualList := sut.ToList()
	assert.Len(t, actualList, 2)
	assert.Contains(t, actualList, SpecificTestStats{
		Name:      "test-1",
		Successes: 2,
		Failures:  1,
	})

	assert.Contains(t, actualList, SpecificTestStats{
		Name:      "test-2",
		Successes: 1,
	})
}

func TestStatsAggregatorToMap(t *testing.T) {
	// GIVEN
	sut := newStatsAggregator()
	sut.AddTestResult("test-1", true)
	sut.AddTestResult("test-2", false)
	// WHEN
	actualMap := sut.ToMap()
	// THEN
	assert.Len(t, actualMap, 2)
	assert.Equal(t, SpecificTestStats{
		Name:      "test-1",
		Successes: 1,
	}, actualMap["test-1"])
	assert.Equal(t, SpecificTestStats{
		Name:     "test-2",
		Failures: 1,
	}, actualMap["test-2"])
}

func TestStatsAggregatorMerge(t *testing.T) {
	// GIVEN
	sut := newStatsAggregator()
	sut.AddTestResult("test-1", true)
	sut.AddTestResult("test-2", false)
	sut.Merge(map[string]SpecificTestStats{
		"test-1": {
			Successes: 1,
			Failures:  1,
		},
		"test-3": {
			Name:      "test-3",
			Successes: 10,
			Failures:  10,
		},
	})
	// WHEN
	actualList := sut.ToList()
	// THEN
	assert.Len(t, actualList, 3)
	assert.Contains(t, actualList, SpecificTestStats{
		Name:      "test-1",
		Successes: 2,
		Failures:  1,
	})

	assert.Contains(t, actualList, SpecificTestStats{
		Name:     "test-2",
		Failures: 1,
	})

	assert.Contains(t, actualList, SpecificTestStats{
		Name:      "test-3",
		Successes: 10,
		Failures:  10,
	})
}
