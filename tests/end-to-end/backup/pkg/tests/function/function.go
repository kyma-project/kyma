package function

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
)

type functionTest struct {
	functionName, testData string
	kubelessClient         *kubeless.Clientset
	coreClient             *kubernetes.Clientset
}

func NewFunctionTest() (functionTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return functionTest{}, err
	}

	kubelessClient, err := kubeless.NewForConfig(restConfig)
	if err != nil {
		return functionTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return functionTest{}, err
	}

	return functionTest{
		kubelessClient: kubelessClient,
		coreClient:     coreClient,
		functionName:   "hello",
		testData:       "test",
	}, nil
}

func (f functionTest) CreateResources(namespace string) {
	_, err := f.createFunction(namespace)
	So(err, ShouldBeNil)
}

func (f functionTest) TestResources(namespace string) {
	err := f.getFunctionPodStatus(namespace, 2*time.Minute)
	So(err, ShouldBeNil)

	host := fmt.Sprintf("http://%s.%s:8080", f.functionName, namespace)
	value, err := f.getFunctionOutput(host, 2*time.Minute)
	So(err, ShouldBeNil)
	So(value, ShouldContainSubstring, f.testData)
}

func (f *functionTest) getFunctionOutput(host string, waitmax time.Duration) (string, error) {
	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(waitmax)
	messages := ""

	for {
		select {
		case <-ticker.C:
			resp, err := http.Post(host, "text/plain", bytes.NewBufferString(f.testData))
			if err != nil {
				messages += fmt.Sprintf("%+v\n", err)
				break
			}
			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
				return string(bodyBytes), nil
			}
			messages += fmt.Sprintf("%+v", err)

		case <-timeout:
			return "", fmt.Errorf("Could not get function output:\n %v", messages)
		}
	}

}

func (f functionTest) createFunction(namespace string) (*kubelessV1.Function, error) {
	function := &kubelessV1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.functionName,
		},
		Spec: kubelessV1.FunctionSpec{
			Handler: "handler.hello",
			Runtime: "nodejs8",
			Function: `module.exports = {
				hello: function(event, context) {
				  return(event.data)
				}
			  }`,
		},
	}
	return f.kubelessClient.KubelessV1beta1().Functions(namespace).Create(function)
}

func (f functionTest) getFunctionPodStatus(namespace string, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-timeout:
			pods, err := f.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + f.functionName})
			if err != nil {
				return err
			}
			return fmt.Errorf("Pod did not start within given time  %v: %+v", waitmax, pods)
		case <-ticker.C:
			pods, err := f.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + f.functionName})
			if err != nil {
				return err
			}
			if len(pods.Items) == 1 {
				pod := pods.Items[0]
				if pod.Status.Phase == corev1.PodRunning {
					return nil
				}
				if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
					return fmt.Errorf("Function in state %v: \n%+v", pod.Status.Phase, pod)
				}
			}
		}
	}
}
