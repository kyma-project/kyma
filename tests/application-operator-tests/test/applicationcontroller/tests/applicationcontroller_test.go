package tests

import (
	"testing"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/applicationcontroller"
)

func TestApplicationOperator(t *testing.T) {
	testSuite := applicationcontroller.NewTestSuite(t)

	t.Run("Application Operator - Application lifecycle test", func(t *testing.T) {
		t.Log("Creating Application without access label")
		testSuite.CreateApplication(t, "", false)

		t.Log("Waiting for Helm release to install")
		testSuite.WaitForReleaseToInstall(t)

		t.Log("Checking if k8s resource deployed")
		testSuite.CheckK8sResourcesDeployed(t)

		t.Log("Checking access label")
		testSuite.CheckAccessLabel(t)

		t.Log("Deleting Application")
		testSuite.DeleteApplication(t)

		t.Log("Waiting for Helm release to delete")
		testSuite.WaitForReleaseToUninstall(t)

		t.Log("Checking if k8s resources removed")
		testSuite.CheckK8sResourceRemoved(t)
	})

	testSuite.CleanUp()
}

func TestApplicationOperator_SkipProvisioning(t *testing.T) {
	testSuite := applicationcontroller.NewTestSuite(t)

	t.Run("Application Operator - skip provisioning test", func(t *testing.T) {
		t.Log("Creating Application without access label")
		testSuite.CreateApplication(t, "", true)

		t.Log("Waiting to ensure release not being installed")
		testSuite.EnsureReleaseNotInstalling(t)

		t.Log("Checking access label")
		testSuite.CheckAccessLabel(t)

		t.Log("Deleting Application")
		testSuite.DeleteApplication(t)
	})

	testSuite.CleanUp()
}
