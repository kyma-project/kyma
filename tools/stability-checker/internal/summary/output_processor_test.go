package summary_test

import (
	"testing"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessorHappyPath(t *testing.T) {
	// GIVEN
	expOutput := []summary.SpecificTestStats{
		{Name: "test-core-core-acceptance", Successes: 1},
		{Name: "test-core-event-bus-tester", Failures: 1},
		{Name: "test-core-environments", Failures: 1},
		{Name: "test-core-kubeless", Failures: 1},
		{Name: "test-core-logging", Failures: 1},
	}

	sut, err := summary.NewOutputProcessor(fixFailureRegexp(), fixSuccessRegexp())
	require.NoError(t, err)

	// WHEN
	results, err := sut.Process(exampleTestOutput)

	// THEN
	require.NoError(t, err)
	assert.Len(t, results, len(expOutput))

	for _, exp := range expOutput {
		assert.Equal(t, exp, results[exp.Name])
	}
}

func TestProcessorOnEmptyInput(t *testing.T) {
	// GIVEN
	sut, err := summary.NewOutputProcessor(fixFailureRegexp(), fixSuccessRegexp())
	require.NoError(t, err)
	// WHEN
	results, err := sut.Process(nil)
	// THEN
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestNewOutputProcessorReturnErrorsOnWrongRegexp(t *testing.T) {
	t.Run("for success", func(t *testing.T) {
		_, err := summary.NewOutputProcessor(fixFailureRegexp(), "abcd")
		assert.EqualError(t, err, "regexp indicating successful tests has to have one capturing group (test name)")

	})

	t.Run("for failure", func(t *testing.T) {
		_, err := summary.NewOutputProcessor("abcd", fixSuccessRegexp())
		assert.EqualError(t, err, "regexp indicating failed tests has to have one capturing group (test name)")
	})
}

func fixSuccessRegexp() string {
	return "Test of '([0-9A-Za-z_-]+)' was successful"
}

func fixFailureRegexp() string {
	return "'([0-9A-Za-z_-]+)' (?:has Failed status?|failed due to too long Running status?|failed due to too long Pending status?|failed with Unknown status?)"
}

var exampleTestOutput = []byte(`
----------------------------
- Testing Kyma...
----------------------------
- Testing Core components...
RUNNING: test-core-logging
ERROR: pods \"test-core-logging\" already exists
RUNNING: test-core-event-bus-tester
ERROR: pods \"test-core-event-bus-tester\" already exists
RUNNING: test-core-kubeless
ERROR: pods \"test-core-kubeless\" already exists
RUNNING: test-core-environments
ERROR: pods \"test-core-environments\" already exists
RUNNING: test-core-core-acceptance
PASSED: test-core-core-acceptance
Error: 10 test(s) failed
\u001b[0m\u001b[1mAll test pods should be terminated. Checking...\u001b[0m
\u001b[32m\u001b[1mOK\u001b[0m
\u001b[0m\u001b[1mTesting 'test-core-core-acceptance'\u001b[0m
Test of 'test-core-core-acceptance' was successful
Logs are not displayed after success
\u001b[0m\u001b[1mEnd of testing 'test-core-core-acceptance'
\u001b[0m
\u001b[0m\u001b[1mTesting 'test-core-environments'\u001b[0m
\u001b[31m'test-core-environments' failed with Unknown status\u001b[0m
\u001b[0m\u001b[1mEnd of testing 'test-core-environments'
\u001b[0m
\u001b[0m\u001b[1mTesting 'test-core-event-bus-tester'\u001b[0m
\u001b[31m'test-core-event-bus-tester' failed due to too long Pending status\u001b[0m
\u001b[0m\u001b[1mEnd of testing 'test-core-event-bus-tester'
\u001b[0m
\u001b[0m\u001b[1mTesting 'test-core-kubeless'\u001b[0m
\u001b[31m'test-core-kubeless' failed due to too long Running status\u001b[0m
\u001b[0m\u001b[1mEnd of testing 'test-core-kubeless'
\u001b[0m
\u001b[0m\u001b[1mTesting 'test-core-logging'\u001b[0m
\u001b[31m'test-core-logging' has Failed status\u001b[0m
\u001b[0m\u001b[1mFetching logs from 'test-core-logging'\u001b[0m
`)
