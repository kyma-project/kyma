package tests

import (
	"testing"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/application"
)

func TestApplicationRelease(t *testing.T) {

	testSuite := application.NewTestSuite(t)

	defer testSuite.Cleanup(t)
	testSuite.Setup(t)

	t.Run("Application tests should succeed", func(t *testing.T) {
		t.Log("Waiting for application to be deployed...")
		testSuite.WaitForApplicationToBeDeployed(t)

		t.Log("Running Application test...")
		testSuite.RunApplicationTests(t)
	})
}
