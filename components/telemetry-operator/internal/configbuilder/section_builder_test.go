package configbuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateOutputSection(t *testing.T) {
	expected := `[OUTPUT]

`
	sut := NewOutputSectionBuilder()
	actual := sut.Build()

	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateFilterSection(t *testing.T) {
	expected := `[FILTER]

`
	sut := NewFilterSectionBuilder()
	actual := sut.Build()

	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}

func TestCreateSectionWithParams(t *testing.T) {
	expected := `[FILTER]
    key1        value1
    key1        value2
    key2        value2
    longer-key1 value1

`
	sut := NewFilterSectionBuilder()
	sut.AddConfigParam("key2", "value2")
	sut.AddConfigParam("key1", "value2")
	sut.AddConfigParam("key1", "value1")
	sut.AddConfigParam("longer-key1", "value1")
	actual := sut.Build()

	require.NotEmpty(t, actual)
	require.Equal(t, expected, actual)
}
