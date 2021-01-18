package tests

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

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

func TestApplicationOperator_LabelOverrides(t *testing.T) {
	testSuite := applicationcontroller.NewTestSuite(t)

	t.Run("Application Operator - label override tests", func(t *testing.T) {
		t.Log("Creating Application with skipVerify=true")
		testSuite.CreateLabeledApplication(t, map[string]string{
			"override.deployment.args.skipVerify": "true",
		})

		t.Log("Waiting for Helm release to install")
		testSuite.WaitForReleaseToInstall(t)

		t.Log("Checking skipVerify=false arg")
		testSuite.AssertRunArgGateway(t, "--skipVerify=true")

		t.Log("Updating application")
		testSuite.UpdateLabeledApplication(t, map[string]string{
			"override.deployment.args.skipVerify": "false",
		})

		t.Log("Check new label exists")
		testSuite.AssertApplicationState(t, func(application *v1alpha1.Application) bool {
			if val, ok := application.Spec.Labels["override.deployment.args.skipVerify"]; ok {
				return val == "false"
			}

			return false
		})

		t.Log("Waiting for Helm release to upgrade")
		testSuite.WaitForReleaseToUpgrade(t)

		t.Log("Checking skipVerify=false arg")
		testSuite.WaitForRunArgGateway(t, "--skipVerify=false")

		t.Log("Deleting Application")
		testSuite.DeleteApplication(t)
	})

	testSuite.CleanUp()
}
