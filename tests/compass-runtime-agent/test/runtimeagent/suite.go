package runtimeagent

import (
	"fmt"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/assertions"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"

	"path/filepath"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	restclient "k8s.io/client-go/rest"
)

const (
	defaultCheckInterval   = 2 * time.Second
	apiServerAccessTimeout = 60 * time.Second
)

type TestSuite struct {
	CompassClient      *compass.Client
	K8sResourceChecker *assertions.K8sResourceChecker
	APIAccessChecker   *assertions.APIAccessChecker

	k8sClient *kubernetes.Clientset

	mockServiceServer *mock.AppMockServer

	config testkit.TestConfig

	mockServiceName string
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

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	appClient, err := v1alpha1.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	serviceClient := k8sClient.Core().Services(config.IntegrationNamespace)
	secretsClient := k8sClient.Core().Secrets(config.IntegrationNamespace)

	nameResolver := applications.NewNameResolver(config.IntegrationNamespace)

	return &TestSuite{
		k8sClient:          k8sClient,
		CompassClient:      compass.NewCompassClient(config.DirectorURL, config.Tenant, config.RuntimeId),
		APIAccessChecker:   assertions.NewAPIAccessChecker(nameResolver),
		K8sResourceChecker: assertions.NewK8sResourceChecker(serviceClient, secretsClient, appClient.Applications(), nameResolver),
		mockServiceServer:  mock.NewAppMockServer(config.MockServicePort),
		config:             config,
		mockServiceName:    config.MockServiceName,
	}, nil
}

func (ts *TestSuite) Setup() error {
	err := ts.waitForAccessToAPIServer()
	if err != nil {
		return errors.Wrap(err, "Error while waiting for access to API server")
	}
	logrus.Infof("Successfully accessed API Server")

	ts.mockServiceServer.Start()

	return nil
}

func (ts *TestSuite) Cleanup() {
	logrus.Infof("Cleaning up after tests...")
	err := ts.mockServiceServer.Kill()
	if err != nil {
		logrus.Errorf("Failed to kill Mock server: %s", err.Error())
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
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", ts.mockServiceName, ts.config.Namespace, ts.config.MockServicePort)
}

func (ts *TestSuite) WaitForProxyInvalidation() {
	// TODO: we should consider introducing some way to invalidate proxy cache
	time.Sleep(time.Duration(ts.config.ProxyInvalidationWaitTime) * time.Second)
}

func (ts *TestSuite) WaitForConfigurationApplication() {
	time.Sleep(time.Duration(ts.config.ConfigApplicationWaitTime) * time.Second)
}
