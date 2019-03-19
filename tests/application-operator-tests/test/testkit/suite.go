package testkit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	testAppName              = "app-ctrl-test-%s"
	defaultCheckInterval     = 2 * time.Second
	installationStartTimeout = 10 * time.Second
	waitBeforeCheck          = 2 * time.Second
)

type TestSuite struct {
	application string

	config     TestConfig
	helmClient HelmClient
	k8sClient  K8sResourcesClient
	k8sChecker *K8sResourceChecker

	installationTimeout time.Duration
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := ReadConfig()
	require.NoError(t, err)

	app := fmt.Sprintf(testAppName, rand.String(5))

	k8sResourcesClient, err := NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := NewHelmClient(config.TillerHost)
	k8sResourcesChecker := NewK8sChecker(k8sResourcesClient, app)

	return &TestSuite{
		application: app,

		config:              config,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		k8sChecker:          k8sResourcesChecker,
		installationTimeout: time.Second * time.Duration(config.ProvisioningTimeout),
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
	err := ts.waitForFunction(defaultCheckInterval, ts.installationTimeout, ts.helmReleaseInstalled)
	require.NoError(t, err, "Received timeout while waiting for release to install")
}

func (ts *TestSuite) WaitForReleaseToUninstall(t *testing.T) {
	err := ts.waitForFunction(defaultCheckInterval, ts.installationTimeout, ts.helmReleaseNotExist)
	require.NoError(t, err, "Received timeout while waiting for release to uninstall")
}

func (ts *TestSuite) EnsureReleaseNotInstalling(t *testing.T) {
	err := ts.shouldLastFor(defaultCheckInterval, installationStartTimeout, ts.helmReleaseNotExist)
	require.NoError(t, err, fmt.Sprintf("Release for %s Application installing when shouldn't", ts.application))
}

func (ts *TestSuite) CheckK8sResourcesDeployed(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.checkK8sResources(t, ts.checkResourceDeployed)
}

func (ts *TestSuite) CheckK8sResourceRemoved(t *testing.T) {
	time.Sleep(waitBeforeCheck)
	ts.k8sChecker.checkK8sResources(t, ts.checkResourceRemoved)
}

func (ts *TestSuite) helmReleaseInstalled() bool {
	status, err := ts.helmClient.CheckReleaseStatus(ts.application)
	return err == nil && status.Info.Status.Code == hapi_4.Status_DEPLOYED
}

func (ts *TestSuite) helmReleaseNotExist() bool {
	exists, err := ts.helmClient.CheckReleaseExistence(ts.application)
	return err == nil && exists == false
}

func (ts *TestSuite) checkResourceDeployed(t *testing.T, resource interface{}, err error, failMessage string) {
	require.NoError(t, err, failMessage)
}

func (ts *TestSuite) checkResourceRemoved(t *testing.T, _ interface{}, err error, failMessage string) {
	require.Error(t, err, failMessage)
	require.True(t, k8serrors.IsNotFound(err), failMessage)
}
