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

	logrus.Infoln("Waiting for test executor to finish...")
	testExecutorStatus := testSuit.WaitForTestExecutorToFinish(t)
	require.NotNil(t, testExecutorStatus.State.Terminated)

	logrus.Infoln("Getting logs from test executor...")
	testSuit.GetTestExecutorLogs(t)

	require.Equal(t, int32(0), testExecutorStatus.State.Terminated.ExitCode, "Test executor exited with code: ", int32(testExecutorStatus.State.Terminated.ExitCode))
}
