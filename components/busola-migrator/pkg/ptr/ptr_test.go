package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	// GIVEN
	testStr := "test"
	expected := &testStr

	// WHEN
	result := String(testStr)

	// THEN
	assert.Equal(t, expected, result)
}

func TestBoolFromString(t *testing.T) {
	t.Run("Valid value true", func(t *testing.T) {
		// GIVEN
		testStr := "true"
		testBool := true
		expected := &testBool

		// WHEN
		result := BoolFromString(testStr)

		// THEN
		assert.Equal(t, expected, result)
	})

	t.Run("Valid value false", func(t *testing.T) {
		// GIVEN
		testStr := "False"
		testBool := false
		expected := &testBool

		// WHEN
		result := BoolFromString(testStr)

		// THEN
		assert.Equal(t, expected, result)
	})

	t.Run("Invalid value", func(t *testing.T) {
		// GIVEN
		testStr := "invalid"

		// WHEN
		result := BoolFromString(testStr)

		// THEN
		assert.Nil(t, result)
	})
}
