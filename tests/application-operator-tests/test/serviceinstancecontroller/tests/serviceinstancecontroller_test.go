package tests

import (
	"testing"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/serviceinstancecontroller"
)

func TestApplicationOperator(t *testing.T) {
	ts := serviceinstancecontroller.NewTestSuite(t)
	t.Log("Creating Namespace")
	ts.Setup(t)
	t.Run("Application Operator - Service Instance lifecycle test", func(t *testing.T) {
		t.Log("Creating Service Instance")
		ts.CreateServiceInstance(t)

		t.Log("Waiting for Helm to install gateway")
		ts.WaitForReleaseToInstall(t)

		t.Log("Checking if k8s resource deployed")
		ts.CheckK8sResourcesDeployed(t)

		t.Log("Deleting Service Instance")
		ts.DeleteServiceInstance(t)

		t.Log("Waiting for Helm to delete gateway")
		ts.WaitForReleaseToUninstall(t)

		t.Log("Checking if k8s resources removed")
		ts.CheckK8sResourceRemoved(t)
	})
	ts.Cleanup()
}
