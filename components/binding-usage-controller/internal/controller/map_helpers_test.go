package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/stretchr/testify/assert"
)

func TestEnsureMapIsInitiated(t *testing.T) {
	t.Run("Not Initiated map", func(t *testing.T) {
		// given
		var notInitiatedMap map[string]string

		// when
		got := controller.EnsureMapIsInitiated(notInitiatedMap)

		// then
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("Already Initiated map", func(t *testing.T) {
		// given
		initiated := map[string]string{
			"key": "val",
		}

		// when
		got := controller.EnsureMapIsInitiated(initiated)

		// then
		assert.Equal(t, got, initiated)
	})
}

func TestMerge(t *testing.T) {
	// GIVEN
	m1 := map[string]string{"a": "1", "b": "2"}
	m2 := map[string]string{"c": "3"}
	// WHEN
	actual, err := controller.Merge(m1, m2)
	// THEN
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "1", "b": "2"}, m1)
	assert.Equal(t, map[string]string{"c": "3"}, m2)
	assert.Equal(t, map[string]string{"a": "1", "b": "2", "c": "3"}, actual)
}

func TestMergeOnConflict(t *testing.T) {
	// GIVEN
	m1 := map[string]string{"a": "1"}
	m2 := map[string]string{"a": "1"}
	// WHEN
	_, err := controller.Merge(m1, m2)
	// THEN
	assert.EqualError(t, err, "Conflict Error for [a]")

}
