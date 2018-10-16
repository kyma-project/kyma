package controllertests

import (
	"github.com/kyma-project/kyma/tests/remote-environment-controller-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

const (
	initialWaitTimeSeconds      = 10 * time.Second
	beforeDeleteWaitTimeSeconds = 20 * time.Second
	retryWaitTimeSeconds        = 5 * time.Second
	retryCount                  = 15
)

func TestRemoteEnvironmentCreation(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost, retryCount, retryWaitTimeSeconds)
	k8sResourcesChecker := testkit.NewK8sCheckerForCreatedResources(k8sResourcesClient, retryCount, retryWaitTimeSeconds)

	t.Run("should create complete RE helm chart when new RE is created", func(t *testing.T) {
		// given
		testReName := "test-create-re"

		// when
		testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName, "")

		// then
		require.NoError(t, err)
		require.NotNil(t, testRe)
		time.Sleep(initialWaitTimeSeconds)

		t.Run("Helm release and k8s resources should exist", func(t *testing.T) {
			releaseAndResourcesShouldExist(t, helmClient, k8sResourcesChecker, testRe.Name)
		})

		t.Run("Access label should be set to RE name", func(t *testing.T) {
			checkAccessLabel(t, k8sResourcesClient, testRe.Name)
		})

		// when
		err = k8sResourcesClient.DeleteRemoteEnvironment(testReName, &v1.DeleteOptions{})

		// then
		require.NoError(t, err)
	})

	t.Run("should overwrite wrong access label", func(t *testing.T) {
		// given
		testReName := "test-create-re-2"

		// when
		testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName, "wrong-access-label")

		// then
		require.NoError(t, err)
		require.NotNil(t, testRe)
		time.Sleep(initialWaitTimeSeconds)

		t.Run("Helm release and k8s resources should exist", func(t *testing.T) {
			releaseAndResourcesShouldExist(t, helmClient, k8sResourcesChecker, testRe.Name)
		})

		t.Run("Access label should be set to RE name", func(t *testing.T) {
			checkAccessLabel(t, k8sResourcesClient, testRe.Name)
		})

		// when
		err = k8sResourcesClient.DeleteRemoteEnvironment(testReName, &v1.DeleteOptions{})

		// then
		require.NoError(t, err)
	})
}

func TestRemoteEnvironmentRemoval(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost, retryCount, retryWaitTimeSeconds)
	k8sResourcesChecker := testkit.NewK8sCheckerForDeletedResources(k8sResourcesClient, retryCount, retryWaitTimeSeconds)

	testReName := "test-delete-re"
	testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName, testReName)
	require.NoError(t, err)
	time.Sleep(beforeDeleteWaitTimeSeconds)

	t.Run("should delete RE helm chart when RE is deleted", func(t *testing.T) {
		// when
		err := k8sResourcesClient.DeleteRemoteEnvironment(testRe.Name, &v1.DeleteOptions{})

		// then
		require.NoError(t, err)
		time.Sleep(initialWaitTimeSeconds)

		// when
		exists, err := helmClient.ExistWhenShouldNot(testRe.Name)

		//then
		require.NoError(t, err)
		require.False(t, exists, "Release %s should not exist but does", testRe.Name)

		k8sResourcesChecker.CheckK8sResources(t, testRe.Name)
	})
}

func releaseAndResourcesShouldExist(t *testing.T, helmClient testkit.HelmClient, k8sChecker testkit.K8sChecker, reName string) {
	// when
	exists, err := helmClient.ExistWhenShould(reName)

	//then
	require.NoError(t, err)
	require.True(t, exists, "Release %s should exist but does not", reName)

	k8sChecker.CheckK8sResources(t, reName)
}

func checkAccessLabel(t *testing.T, k8sClient testkit.K8sResourcesClient, reName string) {
	// when
	receivedRE, err := k8sClient.GetRemoteEnvironment(reName, v1.GetOptions{})

	// then
	require.NoError(t, err)
	require.Equal(t, receivedRE.Name, receivedRE.Spec.AccessLabel)
}
