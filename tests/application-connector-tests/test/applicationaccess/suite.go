package applicationaccess

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit/services"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit/connector"

	types "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	tokenreqversioned "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	tokenreqclient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tokenreq "github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/rand"
	restclient "k8s.io/client-go/rest"
	helmapirelease "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	applicationInstallationTimeout = 240 * time.Second

	defaultCheckInterval       = time.Second * 2
	csrInfoURLRetrievalTimeout = time.Second * 15

	testGroup  = "app-connector-test-group"
	testTenant = "app-connector-test-tenant"
)

type TestSuite struct {
	applicationInstallationTimeout time.Duration
	applicationClient              v1alpha1.ApplicationInterface
	tokenRequestClient             tokenreqclient.TokenRequestInterface

	connectorServiceClient *connector.Client

	isCentral     bool
	skipSSLVerify bool
	defaultGroup  string
	defaultTenant string
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		t.Logf("Failed to read in cluster config, trying with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		require.NoError(t, err)
	}

	applicationClientset, err := versioned.NewForConfig(k8sConfig)
	require.NoError(t, err)

	tokenRequestClientset, err := tokenreqversioned.NewForConfig(k8sConfig)
	require.NoError(t, err)

	var defaultGroup, defaultTenant string

	if config.IsCentral {
		defaultGroup = testGroup
		defaultTenant = testTenant
	}

	return &TestSuite{
		applicationInstallationTimeout: applicationInstallationTimeout,
		applicationClient:              applicationClientset.ApplicationconnectorV1alpha1().Applications(),
		tokenRequestClient:             tokenRequestClientset.ApplicationconnectorV1alpha1().TokenRequests(config.Namespace),
		connectorServiceClient:         connector.NewConnectorClient(config.SkipSSLVerify),
		isCentral:                      config.IsCentral,
		skipSSLVerify:                  config.SkipSSLVerify,
		defaultTenant:                  defaultTenant,
		defaultGroup:                   defaultGroup,
	}
}

func (ts *TestSuite) IsCentral() bool {
	return ts.isCentral
}

func (ts *TestSuite) CleanupApplication(t *testing.T, applicationName string) {
	t.Logf("Cleaning up %s application", applicationName)
	err := ts.applicationClient.Delete(applicationName, &metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) cleanupTokenRequest(t *testing.T, tokenRequestName string) {
	t.Logf("Cleaning up %s token request", tokenRequestName)
	err := ts.tokenRequestClient.Delete(tokenRequestName, &metav1.DeleteOptions{})
	require.NoError(t, err)
}

func (ts *TestSuite) PrepareTestApplication(t *testing.T, namePrefix string) *types.Application {
	name := fmt.Sprintf("%s-%s", namePrefix, rand.String(4))

	application := &types.Application{
		TypeMeta: metav1.TypeMeta{Kind: "Application", APIVersion: types.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: types.ApplicationSpec{
			Services:    []types.Service{},
			Description: "Application deployed by Application Connector Tests",
			Group:       ts.defaultGroup,
			Tenant:      ts.defaultTenant,
		},
	}

	return application
}

func (ts *TestSuite) DeployApplication(t *testing.T, application *types.Application) *types.Application {
	application, err := ts.applicationClient.Create(application)
	require.NoError(t, err)

	return application
}

func (ts *TestSuite) WaitForApplicationToBeDeployed(t *testing.T, applicationName string) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.applicationInstallationTimeout, func() bool {
		t.Log("Waiting for Application to be deployed...")

		app, err := ts.applicationClient.Get(applicationName, metav1.GetOptions{})
		if err != nil {
			return false
		}

		return app.Status.InstallationStatus.Status == helmapirelease.Status_DEPLOYED.String()
	})

	require.NoError(t, err)
}

func (ts *TestSuite) getInfoURL(t *testing.T, application *types.Application) string {
	tokenRequest := &tokenreq.TokenRequest{
		TypeMeta:   metav1.TypeMeta{Kind: "TokenRequest", APIVersion: tokenreq.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: application.Name},
		Context:    tokenreq.ClusterContext{Group: application.Spec.Group, Tenant: application.Spec.Tenant},
		Status:     tokenreq.TokenRequestStatus{ExpireAfter: metav1.Date(2999, time.December, 12, 12, 12, 12, 12, time.Local)},
	}

	tokenRequest, err := ts.tokenRequestClient.Create(tokenRequest)
	require.NoError(t, err)
	defer ts.cleanupTokenRequest(t, tokenRequest.Name)

	tokenRequestName := tokenRequest.Name

	err = testkit.WaitForFunction(defaultCheckInterval, csrInfoURLRetrievalTimeout, func() bool {
		t.Log("Waiting for Info URL in Token Request...")
		tokenRequest, err = ts.tokenRequestClient.Get(tokenRequestName, metav1.GetOptions{})
		return err == nil && tokenRequest.Status.State == "OK"
	})
	require.NoError(t, err)

	return tokenRequest.Status.URL
}

func (ts *TestSuite) EstablishMTLSConnection(t *testing.T, application *types.Application) connector.ApplicationConnection {
	infoURL := ts.getInfoURL(t, application)

	applicationConnection, err := ts.connectorServiceClient.EstablishApplicationConnection(infoURL)
	require.NoError(t, err)

	return applicationConnection
}

func (ts *TestSuite) ShouldAccessApplication(t *testing.T, credentials connector.ApplicationCredentials, urls connector.ManagementInfoURLs) {
	applicationConnectorClient := services.NewApplicationConnectorClient(credentials, urls, ts.skipSSLVerify)

	var apis []services.Service
	var errorResponse *services.ErrorResponse
	err := testkit.Retry(testkit.DefaultRetryConfig, func() (bool, error) {
		apis, errorResponse = applicationConnectorClient.GetAllAPIs(t)
		if errorResponse == nil {
			return false, nil
		}

		if errorResponse.Code == http.StatusServiceUnavailable || errorResponse.Code == http.StatusNotFound {
			t.Logf("Application Registry not ready, received %d status", errorResponse.Code)
			return true, nil
		}

		return false, fmt.Errorf("failed to get all APIs. Status: %d, Error: %s", errorResponse.Code, errorResponse.Error)
	})
	require.NoError(t, err)
	services.RequireNoError(t, errorResponse)
	require.NotNil(t, apis)

	var publishResponse services.PublishResponse
	eventId := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	err = testkit.Retry(testkit.DefaultRetryConfig, func() (bool, error) {
		publishResponse, errorResponse = applicationConnectorClient.SendEvent(t, eventId)
		if errorResponse == nil {
			return false, nil
		}

		if errorResponse.Code == http.StatusServiceUnavailable || errorResponse.Code == http.StatusNotFound {
			t.Logf("Event Service not ready, received %d status", errorResponse.Code)
			return true, nil
		}

		return false, fmt.Errorf("failed to send Event. Status: %d, Error: %s", errorResponse.Code, errorResponse.Error)
	})
	require.NoError(t, err)
	services.RequireNoError(t, errorResponse)

	require.Equal(t, eventId, publishResponse.EventID)
}

func (ts *TestSuite) ShouldFailToAccessApplication(t *testing.T, credentials connector.ApplicationCredentials, urls connector.ManagementInfoURLs, expectedStatus int) {
	applicationConnectorClient := services.NewApplicationConnectorClient(credentials, urls, ts.skipSSLVerify)
	_, errorResponse := applicationConnectorClient.GetAllAPIs(t)
	require.NotNil(t, errorResponse)
	require.Equal(t, expectedStatus, errorResponse.Code)

	eventId := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	_, errorResponse = applicationConnectorClient.SendEvent(t, eventId)
	require.NotNil(t, errorResponse)
	require.Equal(t, expectedStatus, errorResponse.Code)
}

func (ts *TestSuite) ModifyGroupAndTenant(application types.Application, group, tenant string) *types.Application {
	modifiedApp := application.DeepCopy()

	modifiedApp.Spec.Group = group
	modifiedApp.Spec.Tenant = tenant

	return modifiedApp
}
