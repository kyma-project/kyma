package runtimeagent

import (
	"crypto/tls"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/kymaconfig"

	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/clientset"

	app_mapping_v1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/assertions"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/authentication"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/secrets"

	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	rafterapi "github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v1typed "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	defaultCheckInterval   = 2 * time.Second
	apiServerAccessTimeout = 60 * time.Second

	appLabel = "app"
)

type TestSuite struct {
	CompassClient          *compass.Client
	KymaConfigurator       *kymaconfig.KymaConfigurator
	K8sResourceChecker     *assertions.K8sResourceChecker
	ProxyAPIAccessChecker  *assertions.ProxyAPIAccessChecker
	EventsAPIAccessChecker *assertions.EventAPIAccessChecker
	ApplicationCRClient    v1alpha1.ApplicationInterface

	connectorClientSet *clientset.ConnectorClientSet

	k8sClient                *kubernetes.Clientset
	namespaceClient          v1typed.NamespaceInterface
	applicationMappingClient app_mapping_v1alpha1.ApplicationMappingInterface

	mockServiceServer *mock.AppMockServer

	Config testkit.TestConfig

	mockServiceName string
	testPodsLabels  string
}

func NewTestSuite(config testkit.TestConfig) (*TestSuite, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Info("Failed to read in cluster config, trying with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, err
		}
	}

	labelSet := labels.Set{appLabel: config.TestPodAppLabel}
	testPodLabels := labels.SelectorFromSet(labelSet).String()

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	appClient, err := v1alpha1.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	clusterAssetGroupClient, err := newClusterAssetGroupClient(k8sConfig)
	if err != nil {
		return nil, err
	}

	applicationMappingClientset, err := mappingCli.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	appMappingClient := applicationMappingClientset.ApplicationconnectorV1alpha1().ApplicationMappings(config.TestTargetNamespace)

	svcCatalogClient, err := serviceCatalogClient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	scClient := svcCatalogClient.ServiceClasses(config.TestTargetNamespace)
	siClient := svcCatalogClient.ServiceInstances(config.TestTargetNamespace)
	sbClient := svcCatalogClient.ServiceBindings(config.TestTargetNamespace)

	directorClient := compass.NewCompassClient(config.DirectorURL, config.Tenant, config.RuntimeId, config.ScenarioLabel, config.GraphQLLog)
	kymaConfigurator := kymaconfig.NewKymaConfigurator(config.TestTargetNamespace, appMappingClient, scClient, siClient, sbClient)

	return &TestSuite{
		Config:                 config,
		ApplicationCRClient:    appClient.Applications(),
		ProxyAPIAccessChecker:  assertions.NewAPIAccessChecker(config.TestTargetNamespace, kymaConfigurator),
		CompassClient:          directorClient,
		EventsAPIAccessChecker: assertions.NewEventAPIAccessChecker(config.Runtime.EventsURL, directorClient, true),
		K8sResourceChecker:     assertions.NewK8sResourceChecker(appClient.Applications(), clusterAssetGroupClient),

		k8sClient:                k8sClient,
		namespaceClient:          k8sClient.CoreV1().Namespaces(),
		applicationMappingClient: appMappingClient,

		connectorClientSet: clientset.NewConnectorClientSet(clientset.WithSkipTLSVerify(true)),
		mockServiceServer:  mock.NewAppMockServer(config.MockServicePort),
		mockServiceName:    config.MockServiceName,
		testPodsLabels:     testPodLabels,
	}, nil
}

func (ts *TestSuite) Setup() error {
	err := ts.waitForAccessToAPIServer()
	if err != nil {
		return errors.Wrap(err, "Error while waiting for access to API server")
	}
	logrus.Infof("Successfully accessed API Server")

	directorToken, err := ts.getDirectorToken()
	if err != nil {
		return errors.Wrap(err, "Error while getting Director Dex token")
	}
	ts.CompassClient.SetDirectorToken(directorToken)

	ts.mockServiceServer.Start()

	logrus.Infof("Setting up test scenario...")
	err = ts.CompassClient.SetupTestsScenario()
	if err != nil {
		return errors.Wrap(err, "Error while adding tests scenario")
	}

	logrus.Infof("Creating namespace %s...", ts.Config.TestTargetNamespace)
	err = ts.createNamespace()
	if err != nil {
		return errors.Wrap(err, "Error while creating namespace")
	}

	logrus.Infof("Namespace %s created", ts.Config.TestTargetNamespace)

	return nil
}

func (ts *TestSuite) Cleanup() {
	logrus.Infof("Cleaning up after tests...")
	err := ts.mockServiceServer.Kill()
	if err != nil {
		logrus.Errorf("Failed to kill Mock server: %s", err.Error())
	}

	logrus.Infoln("Cleaning up scenarios...")
	err = ts.CompassClient.CleanupTestsScenario()
	if err != nil {
		logrus.Errorf("Failed to remove tests scenario: %s", err.Error())
	}

	logrus.Infoln("Cleaning up Namespace...")
	err = ts.namespaceClient.Delete(ts.Config.TestTargetNamespace, &metav1.DeleteOptions{})
	if err != nil {
		logrus.Errorf("Failed to remove tests Namespace: %s", err.Error())
	}
}

// waitForAccessToAPIServer waits for access to API Server which might be delayed by initialization of Istio sidecar
func (ts *TestSuite) waitForAccessToAPIServer() error {
	err := testkit.WaitForFunction(defaultCheckInterval, apiServerAccessTimeout, func() bool {
		logrus.Info("Trying to access API Server...")
		_, err := ts.k8sClient.ServerVersion()
		if err != nil {
			logrus.Errorf("Failed to access API Server: %s. Retrying in %s", err.Error(), defaultCheckInterval.String())
			return false
		}

		return true
	})

	return err
}

func (ts *TestSuite) createNamespace() error {
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.Config.TestTargetNamespace,
		},
	}

	_, err := ts.namespaceClient.Create(namespace)
	return err
}

func (ts *TestSuite) GenerateCertificateForApplication(t *testing.T, application compass.Application) tls.Certificate {
	oneTimeToken, err := ts.CompassClient.GetOneTimeTokenForApplication(application.ID)
	require.NoError(t, err, "failed to generate one-time token for Application: %s", application.GetContext())

	certificate, err := ts.connectorClientSet.GenerateCertificateForToken(oneTimeToken.Token, oneTimeToken.ConnectorURL)
	require.NoError(t, err, "failed to generate certificate for Application: %s", application.GetContext())

	return certificate
}

func (ts *TestSuite) GetMockServiceURL() string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ts.mockServiceName, ts.Config.CompassNamespace, ts.Config.MockServicePort)
}

func (ts *TestSuite) WaitForApplicationToBeDeployed(t *testing.T, applicationName string) {
	err := testkit.WaitForFunction(defaultCheckInterval, ts.Config.ApplicationInstallationTimeout, func() bool {
		t.Log("Waiting for Application to be deployed...")

		app, err := ts.ApplicationCRClient.Get(applicationName, metav1.GetOptions{})
		if err != nil {
			return false
		}

		logrus.Infof("Application Installation status is: %s", app.Status.InstallationStatus.Status)

		return app.Status.InstallationStatus.Status == "deployed"
	})

	require.NoError(t, err)
}

func (ts *TestSuite) CreateApplicationMapping(appName string) error {
	applicationMapping := &applicationconnectorv1alpha1.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: applicationconnectorv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ts.Config.TestTargetNamespace,
		},
		Spec: applicationconnectorv1alpha1.ApplicationMappingSpec{},
	}

	_, err := ts.applicationMappingClient.Create(applicationMapping)
	return err
}

func (ts *TestSuite) DeleteApplicationMapping(appName string) error {
	return ts.applicationMappingClient.Delete(appName, &metav1.DeleteOptions{})
}

func (ts *TestSuite) WaitForProxyInvalidation() {
	// TODO: we should consider introducing some way to invalidate proxy cache
	time.Sleep(ts.Config.ProxyInvalidationWaitTime)
}

func (ts *TestSuite) WaitForConfigurationApplication() {
	time.Sleep(ts.Config.ConfigApplicationWaitTime)
}

func (ts *TestSuite) getDirectorToken() (string, error) {
	secretInterface := ts.k8sClient.CoreV1().Secrets(ts.Config.DexSecretNamespace)
	secretsRepository := secrets.NewRepository(secretInterface)
	dexSecret, err := secretsRepository.Get(ts.Config.DexSecretName)
	if err != nil {
		return "", err
	}

	return authentication.GetToken(authentication.BuildIdProviderConfig(authentication.EnvConfig{
		Domain:        ts.Config.IdProviderDomain,
		UserEmail:     dexSecret.UserEmail,
		UserPassword:  dexSecret.UserPassword,
		ClientTimeout: ts.Config.IdProviderClientTimeout,
	}))
}

func newClusterAssetGroupClient(config *restclient.Config) (dynamic.ResourceInterface, error) {
	groupVersionResource := rafterapi.GroupVersion.WithResource("clusterassetgroups")

	dynamicClient, e := dynamic.NewForConfig(config)

	if e != nil {
		return nil, e
	}

	return dynamicClient.Resource(groupVersionResource), nil
}
