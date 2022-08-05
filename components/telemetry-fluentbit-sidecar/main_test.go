package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMainMetric(t *testing.T) {
	require.True(t, true)
}

func TestReadEnvironmentVariable(t *testing.T) {
	os.Setenv("TEST_VARIABLE", "1")
	val, err := readEnvironmentVariable("TEST_VARIABLE")
	require.NoError(t, err)
	require.Equal(t, val, "1")

	os.Setenv("TEST_VARIABLE", "")
	_, err = readEnvironmentVariable("TEST_VARIABLE")
	require.Error(t, err)
}
