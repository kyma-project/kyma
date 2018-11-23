package controllertests

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/remote-environment-controller-tests/test/testkit"
	"testing"
)

func TestRemoteEnvironmentController(t *testing.T) {
	testSuite := testkit.NewTestSuite(t)

	t.Run("Remote Environment Controller test", func(t *testing.T) {
		t.Log("Creating Remote Environment")
		testSuite.CreateRemoteEnvironment()

		fmt.Println("Waiting for release install")
		t.Log("Waiting for Helm release to install")
		testSuite.WaitForReleaseToInstall()

		fmt.Println("Waiting for resources")
		t.Log("Checking if k8s resource deployed")
		testSuite.WaitForK8sResourcesToDeploy()

		// TODO - ensure access label

		fmt.Println("Delete RE")
		t.Log("Deleting Remote Environment")
		testSuite.DeleteRemoteEnvironment()

		fmt.Println("Waiting for release to remove")
		t.Log("Waiting for Helm release to delete")
		testSuite.WaitForReleaseToUninstall()

		fmt.Println("Waiting for resources")
		t.Log("Checking if k8s resources removed")
		testSuite.WaitForK8sResourceToDelete()
	})

	testSuite.CleanUp()
}
