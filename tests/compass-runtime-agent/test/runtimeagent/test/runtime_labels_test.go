package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	eventsURLLabelKey  = "runtime_eventServiceUrl"
	consoleURLLabelKey = "runtime_consoleUrl"
)

func TestRuntimeLabeledWithURL(t *testing.T) {

	t.Logf("Fetching Runtime data...")
	runtime, err := testSuite.CompassClient.GetRuntime(testSuite.Config.RuntimeId)
	require.NoError(t, err)
	require.NotNil(t, runtime.Labels)

	t.Logf("Checking if Runtime is labeled with URLs...")
	eventsURLLabelRaw, found := runtime.Labels[eventsURLLabelKey]
	assert.True(t, found)
	consoleURLLabelRaw, found := runtime.Labels[consoleURLLabelKey]
	assert.True(t, found)

	eventsURLLabel, ok := eventsURLLabelRaw.(string)
	assert.True(t, ok)
	consoleURLLabel, ok := consoleURLLabelRaw.(string)
	assert.True(t, ok)

	assert.Equal(t, testSuite.Config.Runtime.EventsURL, eventsURLLabel)
	assert.Equal(t, testSuite.Config.Runtime.ConsoleURL, consoleURLLabel)
}
