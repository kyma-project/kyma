package runtimeagent

import (
	"crypto/tls"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogApi "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	application_mapping_v1alpha "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/clientset"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"
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
	"k8s.io/client-go/util/retry"
	helmapirelease "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	defaultCheckInterval   = 2 * time.Second
	apiServerAccessTimeout = 60 * time.Second
	serviceClassWaitTime   = 30 * time.Second

	appLabel         = "app"
	denierLabelValue = "true"
)

type updatePodFunc func(pod *v1.Pod)

type TestSuite struct {
	CompassClient          *compass.Client
	K8sResourceChecker     *assertions.K8sResourceChecker
	ProxyAPIAccessChecker  *assertions.ProxyAPIAccessChecker
	EventsAPIAccessChecker *assertions.EventAPIAccessChecker
	ApplicationCRClient    v1alpha1.ApplicationInterface

	connectorClientSet *clientset.ConnectorClientSet

	k8sClient       *kubernetes.Clientset
	podClient       v1typed.PodInterface
	namespaceClient v1typed.NamespaceInterface
	nameResolver    *applications.NameResolver

	applicationMappingClient application_mapping_v1alpha.ApplicationMappingInterface
	serviceClassClient       serviceCatalogClient.ServiceClassInterface
	serviceInstanceClient    serviceCatalogClient.ServiceInstanceInterface
	serviceBindingClient     serviceCatalogClient.ServiceBindingInterface

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

	istioClient, err := istioclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	clusterAssetGroupClient, err := newClusterAssetGroupClient(k8sConfig)
	if err != nil {
		return nil, err
	}

	serviceClient := k8sClient.CoreV1().Services(config.IntegrationNamespace)
	secretsClient := k8sClient.CoreV1().Secrets(config.IntegrationNamespace)

	applicationMappingClient, err := application_mapping_v1alpha.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	svcCatalogClient, err := serviceCatalogClient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	nameResolver := applications.NewNameResolver(config.IntegrationNamespace)

	directorClient := compass.NewCompassClient(config.DirectorURL, config.Tenant, config.RuntimeId, config.ScenarioLabel, config.GraphQLLog)

	return &TestSuite{
		Config:                 config,
		ApplicationCRClient:    appClient.Applications(),
		ProxyAPIAccessChecker:  assertions.NewAPIAccessChecker(nameResolver),
		CompassClient:          directorClient,
		EventsAPIAccessChecker: assertions.NewEventAPIAccessChecker(config.Runtime.EventsURL, directorClient, true),
		K8sResourceChecker:     assertions.NewK8sResourceChecker(serviceClient, secretsClient, appClient.Applications(), nameResolver, istioClient, clusterAssetGroupClient, config.IntegrationNamespace),

		k8sClient:                k8sClient,
		podClient:                k8sClient.CoreV1().Pods(config.CompassNamespace),
		namespaceClient:          k8sClient.CoreV1().Namespaces(),
		applicationMappingClient: applicationMappingClient.ApplicationMappings(config.TestTargetNamespace),
		serviceClassClient:       svcCatalogClient.ServiceClasses(config.TestTargetNamespace),
		serviceInstanceClient:    svcCatalogClient.ServiceInstances(config.TestTargetNamespace),

		connectorClientSet: clientset.NewConnectorClientSet(clientset.WithSkipTLSVerify(true)),
		nameResolver:       nameResolver,
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

	err = ts.CompassClient.SetupTestsScenario()
	if err != nil {
		return errors.Wrap(err, "Error while adding tests scenario")
	}

	err = ts.createNamespace()
	if err != nil {
		return errors.Wrap(err, "Error while creating namespace")
	}

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

func (ts *TestSuite) createApplicationMapping(appName string) error {
	applicationMapping := &applicationconnectorv1alpha1.ApplicationMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: ts.Config.TestTargetNamespace,
		},
		Spec: applicationconnectorv1alpha1.ApplicationMappingSpec{},
	}

	_, err := ts.applicationMappingClient.Create(applicationMapping)
	return err
}

func (ts *TestSuite) deleteApplicationMapping(appName string) error {
	return ts.applicationMappingClient.Delete(appName, &metav1.DeleteOptions{})
}

// ProvisionServiceInstances provisions Service Instance and Service Binding for each API package inside the Application
// This results in Application Broker downloading credentials from Director
func (ts *TestSuite) ProvisionServiceInstances(t *testing.T, application compass.Application) []string {
	err := ts.createApplicationMapping(application.Name)
	require.NoError(t, err)

	secretNames := make([]string, 0, len(application.Packages.Data))

	for _, pkg := range application.Packages.Data {
		svcClass := ts.waitForServiceClass(t, pkg.ID)
		t.Logf("Creating Service Instance %s, for %s package, for %s", pkg.ID, pkg.Name, application.GetContext())
		svcInstance, err := ts.createServiceInstance(pkg.ID, svcClass.Name, pkg.Name)
		require.NoError(t, err)
		// TODO: wait for instance to be ready?
		serviceBinding, err := ts.createServiceBinding(pkg.Name, svcInstance)
		require.NoError(t, err)

		// TODO: check Binding status?

		secretNames = append(secretNames, serviceBinding.Name)
	}

	require.NoError(t, err)
}

func (ts *TestSuite) waitForServiceClass(t *testing.T, serviceClassName string) *serviceCatalogApi.ServiceClass {
	t.Logf("Waiting for %s Service Class...", serviceClassName)

	var svcClass *serviceCatalogApi.ServiceClass
	var err error

	err = testkit.WaitForFunction(defaultCheckInterval, serviceClassWaitTime, func() bool {
		svcClass, err = ts.serviceClassClient.Get(serviceClassName, metav1.GetOptions{})
		if err != nil {
			t.Logf("Service Class %s not ready: %s. Retrying until timeout is reached...", serviceClassName, err.Error())
			return false
		}

		return true
	})
	require.NoError(t, err)

	return svcClass
}

func (ts *TestSuite) createServiceInstance(appId, apiPackageId, apiPackageName string) (*v1beta1.ServiceInstance, error) {
	serviceInstance := &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{Name: apiPackageName},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: serviceCatalogApi.PlanReference{
				ServiceClassExternalID: appId,
				ServicePlanExternalID:  apiPackageId,
			},
		},
	}

	return ts.serviceInstanceClient.Create(serviceInstance)
}

func (ts *TestSuite) createServiceBinding(name string, svcInstance *v1beta1.ServiceInstance) (*v1beta1.ServiceBinding, error) {
	serviceBinding := &v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: serviceCatalogApi.LocalObjectReference{Name: svcInstance.Name},
		},
	}

	return ts.serviceBindingClient.Create(serviceBinding)
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

		return app.Status.InstallationStatus.Status == helmapirelease.Status_DEPLOYED.String()
	})

	require.NoError(t, err)
}

func (ts *TestSuite) getResourceNames(t *testing.T, appId string, apiIds ...string) []string {
	serviceNames := make([]string, len(apiIds))

	for i, apiId := range apiIds {
		serviceNames[i] = ts.nameResolver.GetResourceName(appId, apiId)
	}

	return serviceNames
}

func (ts *TestSuite) WaitForProxyInvalidation() {
	// TODO: we should consider introducing some way to invalidate proxy cache
	time.Sleep(ts.Config.ProxyInvalidationWaitTime)
}

func (ts *TestSuite) WaitForConfigurationApplication() {
	time.Sleep(ts.Config.ConfigApplicationWaitTime)
}

func (ts *TestSuite) updatePod(podName string, updateFunc updatePodFunc) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newPod, err := ts.podClient.Get(podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		updateFunc(newPod)
		_, err = ts.podClient.Update(newPod)
		return err
	})
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

func contains(array []string, element string) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}

	return false
}

func newClusterAssetGroupClient(config *restclient.Config) (dynamic.ResourceInterface, error) {
	groupVersionResource := rafterapi.GroupVersion.WithResource("clusterassetgroups")

	dynamicClient, e := dynamic.NewForConfig(config)

	if e != nil {
		return nil, e
	}

	return dynamicClient.Resource(groupVersionResource), nil
}
