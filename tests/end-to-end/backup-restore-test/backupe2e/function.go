package backupe2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils"

	. "github.com/smartystreets/goconvey/convey"

	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type functionTest struct {
	functionName, uuid, namespace string
	kubelessClient                *kubeless.Clientset
	coreClient                    *kubernetes.Clientset
}

func NewFunctionTest() (functionTest, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return functionTest{}, err
	}

	kubelessClient, err := kubeless.NewForConfig(config)
	if err != nil {
		return functionTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return functionTest{}, err
	}
	return functionTest{
		kubelessClient: kubelessClient,
		coreClient:     coreClient,
		functionName:   "hello",
	}, nil
}

func (f functionTest) CreateResources(namespace string) {
	f.namespace = namespace
	Convey("create resources for function test", func(c C) {
		_, err := f.createFunction(namespace, f.functionName)

		So(err, ShouldBeNil)
	})
}

func (f functionTest) TestResources() {
	So(func() string {

		_, err := f.getFunctionPod(f.namespace, f.functionName)
		if err != nil {
			return ""
		}

		return "Running"
	}, utils.ShouldReturnSubstringEventually, "Running", 60*time.Second, 1*time.Second)

	So(func() string {
		host := fmt.Sprintf("http://%s.%s:8080", f.functionName, f.namespace)
		value, err := getFunctionOutput(host, f.namespace, f.functionName, f.uuid)
		if err != nil {
			return "Host not reachable. Retrying"
		}
		return value
	}, utils.ShouldReturnSubstringEventually, f.uuid, 60*time.Second, 1*time.Second)
}

func getFunctionOutput(host, namespace, name, testUUID string) (string, error) {
	resp, err := http.Post(host, "text/plain", bytes.NewBufferString(testUUID))
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "Unable to get response: %v", err
		}
		return string(bodyBytes), err
	}
	return "", err

}

func (f *functionTest) createFunction(namespace, functionName string) (*kubelessV1.Function, error) {
	function := &kubelessV1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: kubelessV1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: functionName,
		},
		Spec: kubelessV1.FunctionSpec{
			Handler: "handler.hello",
			Runtime: "nodejs8",
			Function: `module.exports = {
				hello: function(event, context) {
				  return(event.data)
				}
			  }`,
			ServiceSpec: v1.ServiceSpec{
				Ports: []v1.ServicePort{v1.ServicePort{
					Name:       "http-function-port",
					Port:       8080,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(8080)},
				},
				Type: "ClusterIP",
			},
		},
	}
	return f.kubelessClient.KubelessV1beta1().Functions(namespace).Create(function)
}

func (f *functionTest) getFunctionPod(namespace, functionName string) (v1.Pod, error) {
	pods, err := f.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + functionName})
	if err != nil {
		return v1.Pod{}, err
	}

	for _, pod := range pods.Items {
		isPodRunning := true
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				isPodRunning = false
				break
			}
		}
		if isPodRunning {
			return pod, nil
		}
	}
	return v1.Pod{}, fmt.Errorf("there is no pod ready")
}
