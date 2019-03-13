package tests

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/helmtest/proxy"
)

func TestApplicationGateway(t *testing.T) {

	testSuit := proxy.NewTestSuite(t)
	defer testSuit.Cleanup(t)
	testSuit.Setup(t)

	logrus.Infoln("Waiting for test runner to finish...")
	testRunnerStatus := testSuit.WaitForTestRunnerToFinish(t)
	require.NotNil(t, testRunnerStatus.State.Terminated)

	logrus.Infoln("Getting logs from test runner...")
	testSuit.GetTestRunnerLogs(t)

	require.Equal(t, int32(0), testRunnerStatus.State.Terminated.ExitCode, "Test runner exited with code: ", testRunnerStatus.State.Terminated.ExitCode)
}
