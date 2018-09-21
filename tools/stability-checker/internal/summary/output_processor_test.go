package summary_test

import (
	"testing"
)

func TestProcessorHappyPath(t *testing.T) {
	//// GIVEN
	//passed := "PASSED: ([0-9A-Za-z_-]+)"
	//failed := "FAILED: ([0-9A-Za-z_-]+)"
	//sut, err := analyzer.NewOutputTestProcessor(failed, passed)
	//require.NoError(t, err)
	//// WHEN
	//err := sut.Process(exampleTestOutput)
	//// THEN
	//require.NoError(t, err)
	//
	//assert.Len(t, results, 3)
	//assert.Contains(t, results, analyzer.SpecificTestStats{
	//	Name:      "test-core-environments",
	//	Successes: 1,
	//})
	//assert.Contains(t, results, analyzer.SpecificTestStats{
	//	Name:      "test-core-core-acceptance",
	//	Successes: 1,
	//})
	//assert.Contains(t, results, analyzer.SpecificTestStats{
	//	Name:     "test-core-kubeless",
	//	Failures: 1,
	//})

}

func TestProcessorOnEmptyInput(t *testing.T) {
	// TBD
}

var exampleTestOutput = []byte(`
[7m[2018-09-20T03:39:22.417Z] Output for test id "3ef87a0c-ad44-4e28-9d07-ba343bc5aa56"[0m
 ----------------------------
- Testing Kyma...
----------------------------
- Testing Core components...
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
