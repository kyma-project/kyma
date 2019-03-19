package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/proxy/mock"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/tools"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	defaultCheckInterval           = 2 * time.Second
	appGatewayHealthCheckTimeout   = 15 * time.Second
	accessServiceConnectionTimeout = 45 * time.Second
	apiServerAccessTimeout         = 60 * time.Second
	dnsWaitTime                    = 35 * time.Second

	mockServiceNameFormat     = "%s-gateway-test-mock-service"
	testExecutorPodNameFormat = "%s-tests-test-executor"
)

type TestSuite struct {
	httpClient      *http.Client
	k8sClient       *kubernetes.Clientset
	podClient       corev1.PodInterface
	serviceClient   corev1.ServiceInterface
	config          executor.TestConfig
	appMockServer   *mock.AppMockServer
	mockServiceName string
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := executor.ReadConfig()
	require.NoError(t, err)

	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	require.NoError(t, err)

	appMockServer := mock.NewAppMockServer(config.MockServicePort)

	return &TestSuite{
		httpClient:      &http.Client{},
		k8sClient:       coreClientset,
		podClient:       coreClientset.CoreV1().Pods(config.Namespace),
		serviceClient:   coreClientset.CoreV1().Services(config.Namespace),
		config:          config,
		appMockServer:   appMockServer,
		mockServiceName: fmt.Sprintf(mockServiceNameFormat, config.Application),
	}
}

func (ts *TestSuite) Setup(t *testing.T) {
	ts.WaitForAccessToAPIServer(t)

	ts.appMockServer.Start()
	ts.createMockService(t)

	ts.CheckApplicationGatewayHealth(t)
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	t.Log("Calling cleanup")

	err := ts.appMockServer.Kill()
	assert.NoError(t, err)

	ts.deleteMockService(t)
}

func (ts *TestSuite) ApplicationName() string {
	return ts.config.Application
}

// WaitForAccessToAPIServer waits for access to API Server which might be delayed by initialization of Istio sidecar
func (ts *TestSuite) WaitForAccessToAPIServer(t *testing.T) {
	err := tools.WaitForFunction(defaultCheckInterval, apiServerAccessTimeout, func() bool {
		t.Log("Trying to access API Server...")
		_, err := ts.k8sClient.ServerVersion()
		if err != nil {
			t.Log(err.Error())
			return false
		}

		return true
	})

	require.NoError(t, err)
}

func (ts *TestSuite) CheckApplicationGatewayHealth(t *testing.T) {
	t.Log("Checking application gateway health...")

	err := tools.WaitForFunction(defaultCheckInterval, appGatewayHealthCheckTimeout, func() bool {
		req, err := http.NewRequest(http.MethodGet, ts.proxyURL()+"/v1/health", nil)
		if err != nil {
			return false
		}

		res, err := ts.httpClient.Do(req)
		if err != nil {
			return false
		}

		if res.StatusCode != http.StatusOK {
			return false
		}

		return true
	})

	require.NoError(t, err, "Failed to check health of Application Gateway.")
}

func (ts *TestSuite) CallAccessService(t *testing.T, apiId, path string) *http.Response {
	url := fmt.Sprintf("http://app-%s-%s/%s", ts.config.Application, apiId, path)

	t.Log("Waiting for DNS in Istio Proxy...")
	// Sometimes Istio Proxy has problems with DNS
	// this wait prevents random failures
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

		if resp.StatusCode != http.StatusOK {
			t.Logf("Access service not ready: Invalid response from access service, status: %d.", resp.StatusCode)
			bytes, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			t.Log(string(bytes))

			return false
		}

		return true
	})
	require.NoError(t, err)

	return resp
}

func (ts *TestSuite) proxyURL() string {
	return fmt.Sprintf("http://%s-application-gateway-external-api:8081", ts.config.Application)
}

func (ts *TestSuite) GetMockServiceURL() string {
	return fmt.Sprintf("http://%s:%d/", ts.mockServiceName, ts.config.MockServicePort)
}

func (ts *TestSuite) createMockService(t *testing.T) {
	selectors := map[string]string{
		ts.config.MockSelectorKey: ts.config.MockSelectorValue,
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
	require.NoError(t, err)
}

func (ts *TestSuite) deleteMockService(t *testing.T) {
	err := ts.serviceClient.Delete(ts.mockServiceName, &metav1.DeleteOptions{})
	assert.NoError(t, err)
}

func (ts *TestSuite) AddDenierLabel(t *testing.T, apiId string) {
	pod, err := ts.podClient.Get(fmt.Sprintf(testExecutorPodNameFormat, ts.config.Application), metav1.GetOptions{})
	require.NoError(t, err)

	serviceName := fmt.Sprintf("app-%s-%s", ts.config.Application, apiId)

	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}

	pod.Labels[serviceName] = "true"

	_, err = ts.podClient.Update(pod)
	require.NoError(t, err)
}
