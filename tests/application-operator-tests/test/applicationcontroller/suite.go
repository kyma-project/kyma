package applicationcontroller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"helm.sh/helm/v3/pkg/release"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	testAppName              = "operator-test-%s"
	testGatewayName          = "%s-application-gateway"
	defaultCheckInterval     = 2 * time.Second
	installationStartTimeout = 10 * time.Second
	assessLabelWaitTime      = 15 * time.Second
)

type StateAssertion func(*v1alpha1.Application) bool

type TestSuite struct {
	application string
	gateway     string

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
	gateway := fmt.Sprintf(testGatewayName, app)

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient, err := testkit.NewHelmClient(config.HelmDriver)
	require.NoError(t, err)

	k8sResourcesChecker := testkit.NewAppK8sChecker(k8sResourcesClient, app, !config.GatewayOncePerNamespace)

	return &TestSuite{
		application: app,
		gateway:     gateway,

		config:              config,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		k8sChecker:          k8sResourcesChecker,
		installationTimeout: time.Second * time.Duration(config.InstallationTimeoutSeconds),
	}
}

func (ts *TestSuite) CreateApplication(t *testing.T, accessLabel string, skipInstallation bool) {
	application, err := ts.k8sClient.CreateDummyApplication(context.Background(), ts.application, accessLabel, skipInstallation)
	require.NoError(t, err)
	require.NotNil(t, application)
}

func (ts *TestSuite) CreateLabeledApplication(t *testing.T, labels map[string]string) {
	application, err := ts.k8sClient.CreateLabeledApplication(context.Background(), ts.application, labels)
	require.NoError(t, err)
	require.NotNil(t, application)
}

func (ts *TestSuite) UpdateLabeledApplication(t *testing.T, labels map[string]string) {
	application, err := ts.k8sClient.GetApplication(context.Background(), ts.application, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, application)

	application.Spec.Labels = labels

	updated, err := ts.k8sClient.UpdateApplication(context.Background(), application)
	require.NoError(t, err)
	require.NotNil(t, updated)
}

func (ts *TestSuite) AssertRunArgGateway(t *testing.T, expectedArg string) {
	ts.AssertRunArg(t, ts.gateway, expectedArg)
}

func (ts *TestSuite) WaitForRunArgGateway(t *testing.T, expectedArg string) {
	ts.WaitForRunArg(t, ts.gateway, expectedArg)
}

func (ts *TestSuite) AssertRunArg(t *testing.T, name string, expectedArg string) {
	require.True(t, ts.containsArg(t, name, expectedArg), expectedArg)
}

func (ts *TestSuite) WaitForRunArg(t *testing.T, name string, expectedArg string) {
	err := testkit.WaitForFunction(defaultCheckInterval, assessLabelWaitTime, func() bool {
		return ts.containsArg(t, name, expectedArg)
	})

	require.NoError(t, err)
}

func (ts *TestSuite) containsArg(t *testing.T, name string, expectedArg string) bool {
	deployment, err := ts.k8sClient.GetDeployment(context.Background(), name, metav1.GetOptions{})
	require.NoError(t, err)

	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, arg := range container.Args {
			if arg == expectedArg {
				return true
			}
		}
	}

	return false
}

func (ts *TestSuite) DeleteApplication(t *testing.T) {
	err := ts.k8sClient.DeleteApplication(context.Background(), ts.application, metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) CheckAccessLabel(t *testing.T) {
	var application *v1alpha1.Application

	err := testkit.WaitForFunction(defaultCheckInterval, assessLabelWaitTime, func() bool {
		var err error
		application, err = ts.k8sClient.GetApplication(context.Background(), ts.application, metav1.GetOptions{})
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
	ts.k8sClient.DeleteApplication(context.Background(), ts.application, metav1.DeleteOptions{})
}

func (ts *TestSuite) WaitForReleaseToInstall(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		return ts.helmClient.IsInstalled(ts.application, ts.config.Namespace)
	})
	require.NoError(t, err, "Received timeout while waiting for release to install")
}

func (ts *TestSuite) AssertApplicationState(t *testing.T, assertion StateAssertion) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		application, err := ts.k8sClient.GetApplication(context.Background(), ts.application, metav1.GetOptions{})
		require.NoError(t, err, "Received error while asserting application state")
		return assertion(application)
	})
	require.NoError(t, err, "Received timeout while asserting application state")
}

func (ts *TestSuite) WaitForReleaseToUpgrade(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		status, err := ts.helmClient.CheckReleaseStatus(ts.application, ts.config.Namespace)
		require.NoError(t, err, "Received error while waiting for release to upgrade")
		return status != release.StatusDeployed
	})
	require.NoError(t, err, "Received timeout while waiting for release to upgrade")
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
	return !ts.helmClient.IsInstalled(ts.application, ts.config.Namespace)
}
