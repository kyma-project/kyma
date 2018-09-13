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
	retryWaitTime = 3
	retryCount = 3
)

func TestRemoteEnvironmentCreation(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost, retryCount, retryWaitTime*time.Second)

	testReName := "test-create-re"

	t.Run("should create complete RE helm chart when new RE is created", func(t *testing.T) {
		// when
		testRe, err := k8sResourcesClient.CreateDummyRemoteEnvironment(testReName)

		// then
		require.NoError(t, err)
		time.Sleep(initialWaitTime*time.Second)

		// when
		exists, err := helmClient.ShouldExist(testRe.Name)

		//then
		require.NoError(t, err)
		require.True(t, exists)

		// TODO: check k8s resources

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

		// TODO: check k8s resources
	})
}
