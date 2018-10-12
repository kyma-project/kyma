package gqlschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabels_UnmarshalGQL_Success(t *testing.T) {
	// GIVEN
	l := Labels{}
	fixLabels := map[string]interface{}{
		"fix": "lab",
	}
	expectedLabels := Labels{
		"fix": "lab",
	}

	// WHEN
	err := l.UnmarshalGQL(fixLabels)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, l, expectedLabels)
}

func TestLabels_UnmarshalGQL_Error(t *testing.T) {
	// GIVEN
	l := Labels{}
	fixLabels := map[string]interface{}{
		"fix": 1,
	}

	// WHEN
	err := l.UnmarshalGQL(fixLabels)

	// THEN
	assert.EqualError(t, err, "internal error: while converting labels: given value `1` must be a string")
}

func TestLabels_UnmarshalGQL_CastError(t *testing.T) {
	// GIVEN
	l := Labels{}
	fixLabels := "wrong value"

	// WHEN
	err := l.UnmarshalGQL(fixLabels)

	// THEN
	assert.EqualError(t, err, "internal error: unexpected labels type: string, should be map[string]string")
}
