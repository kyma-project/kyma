package applicationcontroller

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	testAppName              = "operator-test-%s"
	defaultCheckInterval     = 2 * time.Second
	installationStartTimeout = 10 * time.Second
	assessLabelWaitTime      = 15 * time.Second
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

	helmClient, err := testkit.NewHelmClient(config.TillerHost, config.TillerTLSKeyFile, config.TillerTLSCertificateFile, config.TillerTLSSkipVerify)
	require.NoError(t, err)

	k8sResourcesChecker := testkit.NewAppK8sChecker(k8sResourcesClient, app, !config.GatewayOncePerNamespace)

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
	var application *v1alpha1.Application

	err := testkit.WaitForFunction(defaultCheckInterval, assessLabelWaitTime, func() bool {
		var err error
		application, err = ts.k8sClient.GetApplication(ts.application, metav1.GetOptions{})
		if err != nil {
			t.Logf("Failed to get Application while checking Access label: %s", err.Error())
			return false
		}

		if application.Spec.AccessLabel == "" {
			t.Logf("Access label empty, will be retried until timeout is reached")
			return false
		}

		return true
	})

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
	ts.k8sChecker.CheckK8sResourcesDeployed(t)
}

func (ts *TestSuite) CheckK8sResourceRemoved(t *testing.T) {
	ts.k8sChecker.CheckK8sResourceRemoved(t)
}

func (ts *TestSuite) helmReleaseNotExist() bool {
	return !ts.helmClient.IsInstalled(ts.application)
}
