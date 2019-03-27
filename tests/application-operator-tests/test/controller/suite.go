package controller

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"

	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	testAppName              = "operator-test-%s"
	defaultCheckInterval     = 2 * time.Second
	installationStartTimeout = 10 * time.Second
	waitBeforeCheck          = 2 * time.Second
)

type TestSuite struct {
	application string

	config     testkit.TestConfig
	helmClient testkit.HelmClient
	k8sClient  testkit.K8sResourcesClient
	k8sChecker *testkit.K8sResourceChecker

	installationTimeout time.Duration
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	app := fmt.Sprintf(testAppName, rand.String(4))

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost)
	k8sResourcesChecker := testkit.NewK8sChecker(k8sResourcesClient, app)

	return &TestSuite{
		application: app,

		config:              config,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		k8sChecker:          k8sResourcesChecker,
		installationTimeout: time.Second * time.Duration(config.InstallationTimeoutSeconds),
	}
}

func (ts *TestSuite) CreateApplication(t *testing.T, accessLabel string, skipInstallation bool) {
	application, err := ts.k8sClient.CreateDummyApplication(ts.application, accessLabel, skipInstallation)
	require.NoError(t, err)
	require.NotNil(t, application)
}

func (ts *TestSuite) DeleteApplication(t *testing.T) {
	err := ts.k8sClient.DeleteApplication(ts.application, &metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) CheckAccessLabel(t *testing.T) {
	application, err := ts.k8sClient.GetApplication(ts.application, metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, ts.application, application.Spec.AccessLabel)
}

func (ts *TestSuite) CleanUp() {
	// Do not handle error as RE may already be removed
	ts.k8sClient.DeleteApplication(ts.application, &metav1.DeleteOptions{})
}

func (ts *TestSuite) WaitForReleaseToInstall(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		return ts.helmClient.IsInstalled(ts.application)
	})
	require.NoError(t, err, "Received timeout while waiting for release to install")
}

func (ts *TestSuite) WaitForReleaseToUninstall(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, ts.helmReleaseNotExist)
	require.NoError(t, err, "Received timeout while waiting for release to uninstall")
}

func (ts *TestSuite) EnsureReleaseNotInstalling(t *testing.T) {
	err := testkit.ShouldLastFor(defaultCheckInterval, installationStartTimeout, ts.helmReleaseNotExist)
	require.NoError(t, err, fmt.Sprintf("Release for %s Application installing when shouldn't", ts.application))
}

func (ts *TestSuite) CheckK8sResourcesDeployed(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.CheckK8sResources(t, ts.checkResourceDeployed)
}

func (ts *TestSuite) CheckK8sResourceRemoved(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.CheckK8sResources(t, ts.checkResourceRemoved)
}

func (ts *TestSuite) helmReleaseNotExist() bool {
	return !ts.helmClient.IsInstalled(ts.application)
}

func (ts *TestSuite) checkResourceDeployed(t *testing.T, resource interface{}, err error, failMessage string) {
	require.NoError(t, err, failMessage)
}

func (ts *TestSuite) checkResourceRemoved(t *testing.T, _ interface{}, err error, failMessage string) {
	require.Error(t, err, failMessage)
	require.True(t, k8serrors.IsNotFound(err), failMessage)
}
