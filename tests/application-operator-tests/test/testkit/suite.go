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
	testReName               = "re-ctrl-test-%s"
	defaultCheckInterval     = 2 * time.Second
	installationStartTimeout = 10 * time.Second
	waitBeforeCheck          = 2 * time.Second
)

type TestSuite struct {
	t *testing.T

	remoteEnvironment string

	config     TestConfig
	helmClient HelmClient
	k8sClient  K8sResourcesClient
	k8sChecker *K8sResourceChecker

	installationTimeout time.Duration
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

		config:              config,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		k8sChecker:          k8sResourcesChecker,
		installationTimeout: time.Second * time.Duration(config.ProvisioningTimeout),
	}
}

func (ts *TestSuite) CreateRemoteEnvironment(accessLabel string, skipInstallation bool) {
	remoteEnv, err := ts.k8sClient.CreateDummyRemoteEnvironment(ts.remoteEnvironment, accessLabel, skipInstallation)
	require.NoError(ts.t, err)
	require.NotNil(ts.t, remoteEnv)
}

func (ts *TestSuite) DeleteRemoteEnvironment() {
	err := ts.k8sClient.DeleteRemoteEnvironment(ts.remoteEnvironment, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) CheckAccessLabel() {
	remoteEnv, err := ts.k8sClient.GetRemoteEnvironment(ts.remoteEnvironment, metav1.GetOptions{})
	require.NoError(ts.t, err)
	require.Equal(ts.t, ts.remoteEnvironment, remoteEnv.Spec.AccessLabel)
}

func (ts *TestSuite) CleanUp() {
	// Do not handle error as RE may already be removed
	ts.k8sClient.DeleteRemoteEnvironment(ts.remoteEnvironment, &metav1.DeleteOptions{})
}

func (ts *TestSuite) WaitForReleaseToInstall() {
	msg := fmt.Sprintf("Timeout waiting for %s release installation", ts.remoteEnvironment)
	ts.waitForFunction(ts.helmReleaseInstalled, msg, ts.installationTimeout)
}

func (ts *TestSuite) WaitForReleaseToUninstall() {
	msg := fmt.Sprintf("Timeout waiting for %s release to uninstall", ts.remoteEnvironment)
	ts.waitForFunction(ts.helmReleaseNotExist, msg, ts.installationTimeout)
}

func (ts *TestSuite) EnsureReleaseNotInstalling() {
	msg := fmt.Sprintf("Release for %s Remote Environment installing when shouldn't", ts.remoteEnvironment)
	ts.shouldLastFor(ts.helmReleaseNotExist, msg, installationStartTimeout)
}

func (ts *TestSuite) CheckK8sResourcesDeployed() {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.checkK8sResources(ts.checkResourceDeployed)
}

func (ts *TestSuite) CheckK8sResourceRemoved() {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.checkK8sResources(ts.checkResourceRemoved)
}

func (ts *TestSuite) helmReleaseInstalled() bool {
	status, err := ts.helmClient.CheckReleaseStatus(ts.remoteEnvironment)
	return err == nil && status.Info.Status.Code == hapi_4.Status_DEPLOYED
}

func (ts *TestSuite) helmReleaseNotExist() bool {
	exists, err := ts.helmClient.CheckReleaseExistence(ts.remoteEnvironment)
	return err == nil && exists == false
}

func (ts *TestSuite) checkResourceDeployed(resource interface{}, err error, failMessage string) {
	require.NoError(ts.t, err, failMessage)
}

func (ts *TestSuite) checkResourceRemoved(_ interface{}, err error, failMessage string) {
	require.Error(ts.t, err, failMessage)
	require.True(ts.t, k8serrors.IsNotFound(err), failMessage)
}
