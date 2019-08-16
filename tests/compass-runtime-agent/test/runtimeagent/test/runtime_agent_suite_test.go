package test

import (
	"os"
	"testing"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/runtimeagent"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
)

var (
	testSuite *runtimeagent.TestSuite
)

func TestMain(m *testing.M) {
	var exitCode int = 1
	defer os.Exit(exitCode)

	// setup
	logrus.Info("Starting Compass Runtime Agent Test")

	config, err := testkit.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		return
	}

	testSuite, err = runtimeagent.NewTestSuite(config)
	if err != nil {
		logrus.Errorf("Failed to create test suite: %s", err.Error())
		return
	}

	logrus.Info("Setting up...")
	err = testSuite.Setup()
	defer testSuite.Cleanup()
	if err != nil {
		logrus.Errorf("Error while setting up tests: %s", err.Error())
		os.Exit(1)
	}

	// run tests
	logrus.Info("Running tests...")
	exitCode = m.Run()

	// cleanup
	logrus.Info("Tests finished. Exit code: ", exitCode)
}
