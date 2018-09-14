package controllertests

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/kyma-project/kyma/tests/remote-environment-controller-tests/test/testkit"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	initialWaitTime = 5
	retryWaitTime = 2
	retryCount = 3
)

func TestRemoteEnvironmentCreation(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost, retryCount, retryWaitTime*time.Second)
	testReName := "test-create-re"
	k8sResourcesChecker := testkit.NewK8sResourceChecker(testReName, k8sResourcesClient)

	t.Run("should create complete RE helm chart when new RE is created", func(t *testing.T) {
		// when
		testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName)

		// then
		require.NoError(t, err)
		require.NotNil(t, testRe)
		time.Sleep(initialWaitTime*time.Second)

		t.Run("Helm release and k8s resources should exist", func(t *testing.T) {
			// when
			exists, err := helmClient.ShouldExist(testReName)

			//then
			require.NoError(t, err)
			require.True(t, exists)

			k8sResourcesChecker.CheckK8sResources(t, requireNoError, requireNotEmpty)
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

	helmClient := testkit.NewHelmClient(config.TillerHost, retryCount, retryWaitTime*time.Second)

	testReName := "test-delete-re"
	k8sResourcesChecker := testkit.NewK8sResourceChecker(testReName, k8sResourcesClient)

	testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName)
	require.NoError(t, err)
	time.Sleep(initialWaitTime*time.Second)

	t.Run("should delete RE helm chart when RE is deleted", func(t *testing.T) {
		// when
		err := k8sResourcesClient.DeleteRemoteEnvironment(testReName, &v1.DeleteOptions{})

		// then
		require.NoError(t, err)

		// when
		exists, err := helmClient.ShouldNotExist(testRe.Name)

		//then
		require.NoError(t, err)
		require.False(t, exists)

		k8sResourcesChecker.CheckK8sResources(t, requireError, requireEmpty)
	})
}

func requireError(t *testing.T, err error){
	require.Error(t, err)
}

func requireNoError(t *testing.T, err error){
	require.NoError(t, err)
}

func requireNotEmpty(t *testing.T, obj interface{}){
	require.NotEmpty(t, obj)
}

func requireEmpty(t *testing.T, obj interface{}){
	require.Empty(t, obj)
}
