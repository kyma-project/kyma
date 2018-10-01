package controllertests

import (
	"github.com/kyma-project/kyma/tests/remote-environment-controller-tests/test/testkit"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

const (
	initialWaitTimeSeconds = 10 * time.Second
	retryWaitTimeSeconds   = 5 * time.Second
	retryCount             = 6
)

func TestRemoteEnvironmentCreation(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost, retryCount, retryWaitTimeSeconds)
	testReName := "test-create-re"
	k8sResourcesChecker := testkit.NewK8sResourceChecker(testReName, k8sResourcesClient, retryCount, retryWaitTimeSeconds)

	t.Run("should create complete RE helm chart when new RE is created", func(t *testing.T) {
		// when
		testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName)

		// then
		require.NoError(t, err)
		require.NotNil(t, testRe)
		time.Sleep(initialWaitTimeSeconds)

		t.Run("Helm release and k8s resources should exist", func(t *testing.T) {
			// when
			exists, err := helmClient.ExistWhenShould(testRe.Name)

			//then
			require.NoError(t, err)
			require.True(t, exists, "Release %s should exist but does not", testRe.Name)

			k8sResourcesChecker.CheckK8sResources(t, true, requireNoError, requireNotEmpty)
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

	testReName := "test-delete-re"
	k8sResourcesChecker := testkit.NewK8sResourceChecker(testReName, k8sResourcesClient, retryCount, retryWaitTimeSeconds)

	testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName)
	require.NoError(t, err)
	time.Sleep(initialWaitTimeSeconds)

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
		require.False(t, exists, "Release %s should not exist does", testRe.Name)

		k8sResourcesChecker.CheckK8sResources(t, false, requireError, requireEmpty)
	})
}

func requireError(t *testing.T, err error) {
	require.Error(t, err)
	require.True(t, k8serrors.IsNotFound(err))
}

func requireNoError(t *testing.T, err error) {
	require.NoError(t, err)
}

func requireNotEmpty(t *testing.T, obj interface{}) {
	require.NotEmpty(t, obj)
}

func requireEmpty(t *testing.T, obj interface{}) {
	require.Empty(t, obj)
}
