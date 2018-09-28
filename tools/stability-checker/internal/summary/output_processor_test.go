package summary_test

import (
	"testing"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessorHappyPath(t *testing.T) {
	// GIVEN
	sut, err := summary.NewOutputProcessor(fixFailureRegexp(), fixSuccessRegexp())
	require.NoError(t, err)
	// WHEN
	results, err := sut.Process(exampleTestOutput)
	// THEN
	require.NoError(t, err)
	assert.Len(t, results, 5)
	assert.Equal(t, results["test-core-environments"], summary.SpecificTestStats{
		Name:      "test-core-environments",
		Successes: 1,
	})
	assert.Equal(t, results["test-core-core-acceptance"], summary.SpecificTestStats{
		Name:      "test-core-core-acceptance",
		Successes: 1,
	})
	assert.Equal(t, results["test-core-kubeless"], summary.SpecificTestStats{
		Name:     "test-core-kubeless",
		Failures: 1,
	})
	assert.Equal(t, results["test-core-monitoring"], summary.SpecificTestStats{
		Name:     "test-core-monitoring",
		Failures: 1,
	})
	assert.Equal(t, results["remote-environment-controller-tests"], summary.SpecificTestStats{
		Name:     "remote-environment-controller-tests",
		Failures: 1,
	})

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
	return "PASSED: ([0-9A-Za-z_-]+)"
}

func fixFailureRegexp() string {
	return "(?:FAILED: |ERROR: pods \\\\\")([0-9A-Za-z_-]+)"
}

var exampleTestOutput = []byte(`
[7m[2018-09-20T03:39:22.417Z] Output for test id "3ef87a0c-ad44-4e28-9d07-ba343bc5aa56"[0m
 ----------------------------
- Testing Kyma...
----------------------------
- Testing Core components...
RUNNING: test-core-monitoring
ERROR: pods \"test-core-monitoring\" already exists
RUNNING: remote-environment-controller-tests
ERROR: pods \"remote-environment-controller-tests\" is forbidden: exceeded quota: kyma-default, requested: limits.memory=96Mi, used: limits.memory=3038Mi, limited: limits.memory=3Gi
RUNNING: test-core-environments
PASSED: test-core-environments
RUNNING: test-core-core-acceptance
PASSED: test-core-core-acceptance
RUNNING: test-core-kubeless
FAILED: test-core-kubeless, run kubectl logs test-core-kubeless --namespace kyma-system for more info
Error: 1 test(s) failed
[0m[1mAll test pods should be terminated. Checking...[0m
[32m[1mOK[0m
[0m[1mTesting 'connector-service-tests'[0m
Test of 'connector-service-tests' was successful
Logs are not displayed after success
[0m[1mEnd of testing 'connector-service-tests'
...
[0m[1mTesting 'test-core-kubeless'[0m
[31m'test-core-kubeless' has Failed status[0m
[0m[1mFetching logs from 'test-core-kubeless'[0m
2018/09/20 03:35:42 Cleaning up
2018/09/20 03:35:50 Starting test
2018/09/20 03:35:50 Domain Name is: nightly.cluster.kyma.cx
2018/09/20 03:35:51 Deploying test-hello function
2018/09/20 03:35:51 Deploying svc-instance
2018/09/20 03:35:51 Deploying test-event function
2018/09/20 03:35:59 [test-event] Pod: test-event-6cb44dc97-j7hk5: and SBU ID is:
2018/09/20 03:35:59 Publishing event to function test-event
2018/09/20 03:36:00 [test-hello] Pod: test-hello-9dcb8c4c9-xx9nv: and SBU ID is:
2018/09/20 03:36:00 Verifying correct function output for test-hello
2018/09/20 03:36:00 [test-hello] Ingress controller address: '10.0.238.26'
2018/09/20 03:36:00 Unable to publish event:
[0m[1mEnd of testing 'test-core-kubeless'
[0m
[0m[1mTesting 'test-core-monitoring'[0m
Test of 'test-core-monitoring' was successful
Logs are not displayed after success
[0m[1mEnd of testing 'test-core-monitoring'
[0m
[0m[1mTest pods should be marked with label 'helm-chart-test=true'. Checking...[0m
Error from server (NotFound): pods "sample-app-depl-ceusqpqe-66c79b8fdc-f7zqc" not found
[32m[1mOK[0m
[0m[1m
Cleaning up helm test pods[0m
pod "connector-service-tests" deleted
pod "test-api-controller-acceptance" deleted
...
[0m
`)
