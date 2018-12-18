package controllertests

import (
	"testing"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"
)

func TestApplicationOperator(t *testing.T) {
	testSuite := testkit.NewTestSuite(t)

	t.Run("Application Operator - RE lifecycle test", func(t *testing.T) {
		t.Log("Creating Remote Environment without access label")
		testSuite.CreateRemoteEnvironment("", false)

		t.Log("Waiting for Helm release to install")
		testSuite.WaitForReleaseToInstall()

		t.Log("Checking if k8s resource deployed")
		testSuite.CheckK8sResourcesDeployed()

		t.Log("Checking access label")
		testSuite.CheckAccessLabel()

		t.Log("Deleting Remote Environment")
		testSuite.DeleteRemoteEnvironment()

		t.Log("Waiting for Helm release to delete")
		testSuite.WaitForReleaseToUninstall()

		t.Log("Checking if k8s resources removed")
		testSuite.CheckK8sResourceRemoved()
	})

	testSuite.CleanUp()
}

func TestApplicationOperator_SkipProvisioning(t *testing.T) {
	testSuite := testkit.NewTestSuite(t)

	t.Run("Application Operator - skip provisioning test", func(t *testing.T) {
		t.Log("Creating Remote Environment without access label")
		testSuite.CreateRemoteEnvironment("", true)

		t.Log("Waiting to ensure release not being installed")
		testSuite.EnsureReleaseNotInstalling()

		t.Log("Checking access label")
		testSuite.CheckAccessLabel()

		t.Log("Deleting Remote Environment")
		testSuite.DeleteRemoteEnvironment()
	})

	testSuite.CleanUp()
}
