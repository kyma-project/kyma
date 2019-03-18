package application

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/testkit"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	testAppName              = "app-ctrl-app-test-%s"
	defaultCheckInterval     = 2 * time.Second
	installationStartTimeout = 10 * time.Second
	waitBeforeCheck          = 2 * time.Second
)

type TestSuite struct {
	application string

	config     testkit.TestConfig
	helmClient testkit.HelmClient
	k8sClient  testkit.K8sResourcesClient

	installationTimeout time.Duration
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	app := fmt.Sprintf(testAppName, rand.String(5))

	k8sResourcesClient, err := testkit.NewK8sResourcesClient(config.Namespace)
	require.NoError(t, err)

	helmClient := testkit.NewHelmClient(config.TillerHost)

	return &TestSuite{
		application: app,

		config:              config,
		helmClient:          helmClient,
		k8sClient:           k8sResourcesClient,
		installationTimeout: time.Second * time.Duration(config.InstallationTimeout),
	}
}

func (ts *TestSuite) Setup(t *testing.T) {
	log.Infof("Creating %s Application", ts.application)
	_, err := ts.k8sClient.CreateDummyApplication(ts.application, ts.application, false)
	require.NoError(t, err)

	ts.WaitForReleaseToInstall(t)
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	log.Info("Cleaning up...")
	err := ts.k8sClient.DeleteApplication(ts.application, &metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) RunApplicationTests(t *testing.T) {
	// TODO - run application Helm test

	log.Info("Running application tests")
}

func (ts *TestSuite) WaitForReleaseToInstall(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.installationTimeout, func() bool {
		return ts.helmClient.IsInstalled(ts.application)
	})
	require.NoError(t, err, "Received timeout while waiting for release to install")
}
