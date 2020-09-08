package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/helmtest"

	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/tools"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	mockSelectorKey    = "app"
	mockSelectorFormat = "%s-mock-service"

	testExecutorContainerPortName = "http-port"

	defaultCheckInterval = 2 * time.Second
	testExecutorTimeout  = 600 * time.Second

	applicationEnv     = "APPLICATION"
	namespaceEnv       = "NAMESPACE"
	selectorKeyEnv     = "SELECTOR_KEY"
	selectorValueEnv   = "SELECTOR_VALUE"
	mockServicePortEnv = "MOCK_SERVICE_PORT"
)

type TestSuite struct {
	httpClient *http.Client
	podClient  corev1.PodInterface
	config     helmtest.TestConfig

	testExecutorName  string
	mockSelectorValue string
}

func NewTestSuite(t *testing.T) *TestSuite {
	config, err := helmtest.ReadConfig()
	require.NoError(t, err)

	k8sConfig, err := restclient.InClusterConfig()
	require.NoError(t, err)

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	require.NoError(t, err)

	return &TestSuite{
		httpClient:        &http.Client{},
		podClient:         coreClientset.CoreV1().Pods(config.Namespace),
		config:            config,
		testExecutorName:  fmt.Sprintf("%s-tests-test-executor", config.Application),
		mockSelectorValue: fmt.Sprintf(mockSelectorFormat, config.Application),
	}
}

func (ts *TestSuite) Setup(t *testing.T) {
	t.Log("Creating Test Executor pod.")
	ts.CreateTestExecutorPod(t)
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	t.Log("Cleaning up...")
	ts.DeleteTestExecutorPod(t)
}

func (ts *TestSuite) CreateTestExecutorPod(t *testing.T) {
	testExecutorPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ts.testExecutorName,
			Namespace: ts.config.Namespace,
			Labels: map[string]string{
				mockSelectorKey: ts.mockSelectorValue,
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: ts.config.ServiceAccountName,
			Containers: []v1.Container{
				{
					Name:  ts.testExecutorName,
					Image: ts.config.TestExecutorImage,
					Env: []v1.EnvVar{
						{Name: applicationEnv, Value: ts.config.Application},
						{Name: namespaceEnv, Value: ts.config.Namespace},
						{Name: selectorKeyEnv, Value: mockSelectorKey},
						{Name: selectorValueEnv, Value: ts.mockSelectorValue},
						{Name: mockServicePortEnv, Value: strconv.Itoa(ts.config.MockServicePort)},
					},
					Ports: []v1.ContainerPort{
						{
							Name:          testExecutorContainerPortName,
							HostPort:      int32(ts.config.MockServicePort),
							ContainerPort: int32(ts.config.MockServicePort),
							Protocol:      v1.ProtocolTCP,
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	_, err := ts.podClient.Create(testExecutorPod)
	require.NoError(t, err)
}

func (ts *TestSuite) WaitForTestExecutorToFinish(t *testing.T) *v1.ContainerStatus {
	var testsStatus *v1.ContainerStatus

	err := tools.WaitForFunction(defaultCheckInterval, testExecutorTimeout, func() bool {
		var finished bool

		testExecutorPod, err := ts.podClient.Get(ts.testExecutorName, metav1.GetOptions{})
		if err != nil {
			return false
		}

		testsStatus, finished = ts.isPodFinished(testExecutorPod)

		return finished
	})

	require.NoError(t, err)
	require.NotEmpty(t, testsStatus)

	return testsStatus
}

func (ts *TestSuite) isPodFinished(pod *v1.Pod) (*v1.ContainerStatus, bool) {
	for _, c := range pod.Status.ContainerStatuses {
		if c.Name == ts.testExecutorName {
			if c.State.Terminated != nil {
				return &c, true
			}
		}
	}

	return nil, false
}

func (ts *TestSuite) GetTestExecutorLogs(t *testing.T) {
	req := ts.podClient.GetLogs(ts.testExecutorName, &v1.PodLogOptions{Container: ts.testExecutorName})

	reader, err := req.Stream()
	require.NoError(t, err)

	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	require.NoError(t, err)

	t.Log("--------------------------------------------Logs from test executor--------------------------------------------")
	t.Log(string(bytes))
	t.Log("--------------------------------------------End of logs from test executor--------------------------------------------")
}

func (ts *TestSuite) DeleteTestExecutorPod(t *testing.T) {
	err := ts.podClient.Delete(ts.testExecutorName, &metav1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			t.Logf("Failed to delete test executor: %s", err.Error())
			t.FailNow()
		}
	}
}
