package test

import (
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"

	"github.com/sirupsen/logrus"
)

var (
	testSuite *runtimeagent.TestSuite
)

func TestMain(m *testing.M) {
	logrus.Info("Starting Compass Runtime Agent Test")

	exitCode := runTests(m)

	logrus.Info("Tests finished. Exit code: ", exitCode)
	os.Exit(exitCode)
}

func runTests(m *testing.M) int {
	// setup
	config, err := testkit.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		return 1
	}

	testSuite, err = runtimeagent.NewTestSuite(config)
	if err != nil {
		logrus.Errorf("Failed to create test suite: %s", err.Error())
		return 1
	}

	logrus.Info("Setting up...")
	err = testSuite.Setup()
	defer testSuite.Cleanup()
	if err != nil {
		logrus.Errorf("Error while setting up tests: %s", err.Error())
		return 1
	}

	// run tests
	logrus.Info("Running tests...")
	return m.Run()
}
