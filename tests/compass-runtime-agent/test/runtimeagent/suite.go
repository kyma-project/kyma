package runtimeagent

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/applications"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/assertions"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/compass"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/tools"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/mock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"path/filepath"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	restclient "k8s.io/client-go/rest"
)

const (
	defaultCheckInterval           = 2 * time.Second
	appGatewayHealthCheckTimeout   = 15 * time.Second
	accessServiceConnectionTimeout = 90 * time.Second
	apiServerAccessTimeout         = 60 * time.Second
	dnsWaitTime                    = 15 * time.Second

	mockServiceNameFormat = "runtime-agent-mock-service-%s"
)

type TestSuite struct {
	CompassClient      *compass.Client
	K8sResourceChecker *assertions.K8sResourceChecker
	APIAccessChecker   *assertions.APIAccessChecker

	k8sClient     *kubernetes.Clientset
	serviceClient corev1.ServiceInterface

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

	serviceClient := k8sClient.Core().Services(config.Namespace)
	secretsClient := k8sClient.Core().Secrets(config.Namespace)

	nameResolver := applications.NewNameResolver(config.Namespace)

	return &TestSuite{
		k8sClient:          k8sClient,
		serviceClient:      serviceClient,
		CompassClient:      compass.NewCompassClient(config.DirectorURL, config.Tenant, ""), //TODO - runtime Id
		APIAccessChecker:   assertions.NewAPIAccessChecker(),
		K8sResourceChecker: assertions.NewK8sResourceChecker(serviceClient, secretsClient, appClient.Applications(), nameResolver),
		mockServiceServer:  mock.NewAppMockServer(config.MockServicePort),
		config:             config,
		mockServiceName:    fmt.Sprintf(mockServiceNameFormat, rand.String(4)),
	}, nil
}

func (ts *TestSuite) Setup() error {
	err := ts.waitForAccessToAPIServer()
	if err != nil {
		return errors.Wrap(err, "Error while waiting for access to API server")
	}

	ts.mockServiceServer.Start()
	err = ts.createMockService()
	if err != nil {
		return errors.Wrap(err, "Error while creating service for mock server")
	}

	//ts.CheckApplicationGatewayHealth(t) // TODO - might do it while checking if runtime applied config
	return nil
}

func (ts *TestSuite) Cleanup() {
	err := ts.mockServiceServer.Kill()
	if err != nil {
		logrus.Errorf("Failed to kill Mock server: %s", err.Error())
	}

	err = ts.deleteMockService()
	if err != nil {
		logrus.Errorf("Failed to delete mock service: %s", err.Error())
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

func (ts *TestSuite) createMockService() error {
	selectors := map[string]string{
		ts.config.MockServiceSelectorKey: ts.config.MockServiceSelectorValue,
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ts.mockServiceName,
			Namespace: ts.config.Namespace,
			Labels:    selectors,
		},
		Spec: v1.ServiceSpec{
			Selector: selectors,
			Ports: []v1.ServicePort{
				{Port: ts.config.MockServicePort, Name: "http-port"},
			},
		},
	}

	_, err := ts.serviceClient.Create(service)
	if err != nil {
		return errors.Wrap(err, "Failed to create mock service")
	}

	return nil
}

func (ts *TestSuite) deleteMockService() error {
	return ts.serviceClient.Delete(ts.mockServiceName, &metav1.DeleteOptions{})
}

//func (ts *TestSuite) proxyURL() string {
//	return fmt.Sprintf("http://%s-application-gateway-external-api:8081", ts.config.Application)
//}

func (ts *TestSuite) GetMockServiceURL() string {
	return fmt.Sprintf("http://%s:%d", ts.mockServiceName, ts.config.MockServicePort)
}

func (ts *TestSuite) CallAccessService(t *testing.T, application, apiId, path string) *http.Response {
	url := fmt.Sprintf("http://app-%s-%s/%s", application, apiId, path) // TODO: it might not be that simple as there might be problems with name length

	t.Log("Waiting for DNS in Istio Proxy...")
	// Wait for Istio Pilot to propagate DNS
	time.Sleep(dnsWaitTime)

	var resp *http.Response

	err := tools.WaitForFunction(defaultCheckInterval, accessServiceConnectionTimeout, func() bool {
		t.Logf("Accessing proxy at: %s", url)
		var err error

		resp, err = http.Get(url)
		if err != nil {
			t.Logf("Access service not ready: %s", err.Error())
			return false
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusServiceUnavailable {
			t.Logf("Invalid response from access service, status: %d.", resp.StatusCode)
			bytes, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			t.Log(string(bytes))
			t.Logf("Access service is not ready. Retrying.")
		}

		return true
	})
	require.NoError(t, err)

	return resp
}
