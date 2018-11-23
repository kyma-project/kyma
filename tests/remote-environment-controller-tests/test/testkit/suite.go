package testkit

import (
	"fmt"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
	"testing"
	"time"
)

const (
	testReName = "re-ctrl-test-%s"
)

type TestSuite struct {
	t *testing.T

	remoteEnvironment string

	config     TestConfig
	helmClient HelmClient
	k8sClient  K8sResourcesClient
	k8sChecker *K8sResourceChecker

	installationTimeout       time.Duration
	installationCheckInterval time.Duration
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := ReadConfig()
	require.NoError(t, err)

	re := fmt.Sprintf(testReName, rand.String(5))

	k8sResourcesClient, err := NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := NewHelmClient(config.TillerHost)
	k8sResourcesChecker := NewK8sChecker(k8sResourcesClient, re)

	return &TestSuite{
		t:                 t,
		remoteEnvironment: re,

		config:     config,
		helmClient: helmClient,
		k8sClient:  k8sResourcesClient,
		k8sChecker: k8sResourcesChecker,

		installationTimeout:       180 * time.Second,
		installationCheckInterval: 2 * time.Second,
	}
}

func (ts *TestSuite) CreateRemoteEnvironment() {
	remoteEnv, err := ts.k8sClient.CreateDummyRemoteEnvironment(ts.remoteEnvironment, "", false)
	require.NoError(ts.t, err)
	require.NotNil(ts.t, remoteEnv)
}

func (ts *TestSuite) DeleteRemoteEnvironment() {
	err := ts.k8sClient.DeleteRemoteEnvironment(ts.remoteEnvironment, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) CleanUp() {
	// Do not handle error as RE may already be removed
	ts.k8sClient.DeleteRemoteEnvironment(ts.remoteEnvironment, &metav1.DeleteOptions{})
}

func (ts *TestSuite) WaitForReleaseToInstall() {
	msg := fmt.Sprintf("Timeout waiting for %s release installation", ts.remoteEnvironment)
	ts.waitForFunction(ts.helmReleaseInstalled, msg)
}

func (ts *TestSuite) WaitForReleaseToUninstall() {
	msg := fmt.Sprintf("Timeout waiting for %s release to uninstall", ts.remoteEnvironment)
	ts.waitForFunction(ts.helmReleaseRemoved, msg)
}

func (ts *TestSuite) WaitForK8sResourcesToDeploy() {
	ts.waitForFunctions(
		ts.k8sChecker.getResourceCheckFunctions(checkResourceDeployed),
	)
}

func (ts *TestSuite) WaitForK8sResourceToDelete() {
	ts.waitForFunctions(
		ts.k8sChecker.getResourceCheckFunctions(checkResourceRemoved),
	)
}

func (ts *TestSuite) helmReleaseInstalled() bool {
	status, err := ts.helmClient.CheckReleaseStatus(ts.remoteEnvironment)
	return err == nil && status.Info.Status.Code == hapi_4.Status_DEPLOYED
}

func (ts *TestSuite) helmReleaseRemoved() bool {
	exists, err := ts.helmClient.CheckReleaseExistence(ts.remoteEnvironment)
	return err == nil && exists == false
}

func checkResourceDeployed(resource interface{}, err error) func() bool {
	return func() bool {
		return err == nil && resource != nil
	}
}

func checkResourceRemoved(resource interface{}, err error) func() bool {
	return func() bool {
		fmt.Println(err)
		fmt.Println(k8serrors.IsNotFound(err))
		return err != nil && k8serrors.IsNotFound(err)
	}
}
