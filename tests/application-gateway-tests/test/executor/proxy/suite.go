package proxy

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/proxy/mock"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/tools"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	defaultCheckInterval           = 2 * time.Second
	appGatewayHealthCheckTimeout   = 15 * time.Second
	accessServiceConnectionTimeout = 120 * time.Second
	apiServerAccessTimeout         = 60 * time.Second

	mockServiceNameFormat   = "%s-gateway-test-mock-service"
	testRunnerPodNameFormat = "%s-tests-test-runner"
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
	log.Infoln("Calling cleanup")

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
		log.Infoln("Trying to access API Server...")
		_, err := ts.k8sClient.ServerVersion()
		if err != nil {
			log.Errorf(err.Error())
			return false
		}

		return true
	})

	require.NoError(t, err)
}

func (ts *TestSuite) CheckApplicationGatewayHealth(t *testing.T) {
	log.Infoln("Checking application gateway health...")

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

	var resp *http.Response

	err := tools.WaitForFunction(defaultCheckInterval, accessServiceConnectionTimeout, func() bool {
		log.Infoln("Accessing proxy at: ", url)
		var err error

		resp, err = http.Get(url)
		if err != nil {
			log.Errorf("Access service not ready: %s", err.Error())
			return false
		}

		if resp.StatusCode != http.StatusOK {
			log.Errorf("Access service not ready: Invalid response from access service, status: %d.", resp.StatusCode)
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
	pod, err := ts.podClient.Get(fmt.Sprintf(testRunnerPodNameFormat, ts.config.Application), metav1.GetOptions{})
	require.NoError(t, err)

	serviceName := fmt.Sprintf("app-%s-%s", ts.config.Application, apiId)

	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}

	pod.Labels[serviceName] = "true"

	_, err = ts.podClient.Update(pod)
	require.NoError(t, err)
}
