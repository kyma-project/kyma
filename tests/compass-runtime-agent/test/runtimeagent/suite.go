package runtimeagent

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/secrets"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	istioclient "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	scheme "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/assertions"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/authentication"
	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	appLabel         = "app"
	denierLabelValue = "true"
)

type updatePodFunc func(pod *v1.Pod)

type TestSuite struct {
	CompassClient       *compass.Client
	K8sResourceChecker  *assertions.K8sResourceChecker
	APIAccessChecker    *assertions.APIAccessChecker
	ApplicationCRClient v1alpha1.ApplicationInterface

	k8sClient    *kubernetes.Clientset
	podClient    v1typed.PodInterface
	nameResolver *applications.NameResolver

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

	clusterDocsTopicClient, err := newClusterDocsTopicClient(k8sConfig)
	if err != nil {
		return nil, err
	}

	serviceClient := k8sClient.CoreV1().Services(config.IntegrationNamespace)
	secretsClient := k8sClient.CoreV1().Secrets(config.IntegrationNamespace)

	nameResolver := applications.NewNameResolver(config.IntegrationNamespace)

	return &TestSuite{
		k8sClient:           k8sClient,
		podClient:           k8sClient.CoreV1().Pods(config.Namespace),
		ApplicationCRClient: appClient.Applications(),
		nameResolver:        nameResolver,
		CompassClient:       compass.NewCompassClient(config.DirectorURL, config.Tenant, config.RuntimeId, config.ScenarioLabel, config.GraphQLLog),
		APIAccessChecker:    assertions.NewAPIAccessChecker(nameResolver),
		K8sResourceChecker:  assertions.NewK8sResourceChecker(serviceClient, secretsClient, appClient.Applications(), nameResolver, istioClient, clusterDocsTopicClient, config.IntegrationNamespace),
		mockServiceServer:   mock.NewAppMockServer(config.MockServicePort),
		Config:              config,
		mockServiceName:     config.MockServiceName,
		testPodsLabels:      testPodLabels,
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

func (ts *TestSuite) GetMockServiceURL() string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ts.mockServiceName, ts.Config.Namespace, ts.Config.MockServicePort)
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

func (ts *TestSuite) AddDenierLabels(t *testing.T, appId string, apiIds ...string) {
	pod := ts.getTestPod(t)
	serviceNames := ts.getResourceNames(t, appId, apiIds...)

	updateFunc := func(pod *v1.Pod) {
		if pod.Labels == nil {
			pod.Labels = map[string]string{}
		}

		for _, svcName := range serviceNames {
			pod.Labels[svcName] = denierLabelValue
		}
	}

	err := ts.updatePod(pod.Name, updateFunc)
	require.NoError(t, err)
}

func (ts *TestSuite) RemoveDenierLabels(t *testing.T, appId string, apiIds ...string) {
	pod := ts.getTestPod(t)
	labelsToRemove := ts.getResourceNames(t, appId, apiIds...)

	updateFunc := func(pod *v1.Pod) {
		newLabels := map[string]string{}

		for name, label := range pod.Labels {
			if !contains(labelsToRemove, name) {
				newLabels[name] = label
			}
		}

		pod.Labels = newLabels
	}

	err := ts.updatePod(pod.Name, updateFunc)
	require.NoError(t, err)
}

func (ts *TestSuite) getTestPod(t *testing.T) v1.Pod {
	testPods, err := ts.podClient.List(metav1.ListOptions{LabelSelector: ts.testPodsLabels})
	require.NoError(t, err)
	assert.True(t, len(testPods.Items) != 0)

	if len(testPods.Items) > 1 {
		return getYoungestPod(testPods.Items)
	}

	return testPods.Items[0]
}

func getYoungestPod(pods []v1.Pod) v1.Pod {
	youngestPod := pods[0]
	for _, p := range pods {
		if p.CreationTimestamp.Unix() > youngestPod.CreationTimestamp.Unix() {
			youngestPod = p
		}
	}

	return youngestPod
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

func newClusterDocsTopicClient(config *restclient.Config) (dynamic.ResourceInterface, error) {
	groupVersionResource := schema.GroupVersionResource{
		Version:  scheme.GroupVersion.Version,
		Group:    scheme.GroupVersion.Group,
		Resource: "clusterdocstopics",
	}

	dynamicClient, e := dynamic.NewForConfig(config)

	if e != nil {
		return nil, e
	}

	return dynamicClient.Resource(groupVersionResource), nil
}
