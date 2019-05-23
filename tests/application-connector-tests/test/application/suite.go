package application

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit/connector"

	types "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"
	restclient "k8s.io/client-go/rest"
	helmapirelease "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	defaultCheckInterval = time.Second * 2
	testAppNameFormat    = "app-conn-test-%s"
)

type TestSuite struct {
	applicationName                string
	applicationInstallationTimeout time.Duration
	applicationClient              v1alpha1.ApplicationInterface

	application *types.Application

	connectorServiceClient *connector.Client

	applicationConnection      connector.ApplicationConnection
	applicationConnectorClient *http.Client
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	app := fmt.Sprintf(testAppNameFormat, rand.String(4))

	// TODO - try to get local config if cluster not found
	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	//coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	//require.NoError(t, err)

	applicationClientset, err := versioned.NewForConfig(k8sConfig)
	require.NoError(t, err)

	return &TestSuite{
		applicationName:   app,
		applicationClient: applicationClientset.ApplicationconnectorV1alpha1().Applications(),

		connectorServiceClient: connector.NewConnectorClient(config.ConnectorInternalAPIURL),
	}
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	t.Log("Cleaning up...")
	err := ts.applicationClient.Delete(ts.applicationName, &metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) CreateTestApplication(t *testing.T) {
	application := &types.Application{
		ObjectMeta: v1.ObjectMeta{
			Name: ts.applicationName,
		},
		Spec: types.ApplicationSpec{
			Description: "Application deployed by Application Connector Tests",
		},
	}

	t.Logf("Creating %s application...", application.Name)
	application, err := ts.applicationClient.Create(application)
	require.NoError(t, err)
	t.Logf("Application created")

	ts.application = application
}

func (ts *TestSuite) EstablishMTLSConnection(t *testing.T) {
	applicationConnection, err := ts.connectorServiceClient.EstablishApplicationConnection(ts.application)
	require.NoError(t, err)

	ts.applicationConnection = applicationConnection
	ts.applicationConnectorClient = applicationConnection.NewMTLSClient()
}

func (ts *TestSuite) WaitForApplicationToBeDeployed(t *testing.T) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.applicationInstallationTimeout, func() bool {
		app, err := ts.applicationClient.Get(ts.applicationName, metav1.GetOptions{})
		if err != nil {
			return false
		}

		return app.Status.InstallationStatus.Status == helmapirelease.Status_DEPLOYED.String()
	})

	require.NoError(t, err)
}
