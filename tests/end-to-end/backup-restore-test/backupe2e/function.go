package backupe2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/smartystreets/goconvey/convey"

	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type functionTest struct {
	functionName, uuid string
	kubelessClient     *kubeless.Clientset
	coreClient         *kubernetes.Clientset
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
		uuid:           uuid.New().String(),
	}, nil
}

func (f functionTest) CreateResources(namespace string) {
	Convey("create resources for function test", func(c C) {
		_, err := f.createFunction(namespace)

		So(err, ShouldBeNil)
	})
}

func (f functionTest) TestResources(namespace string) {
	err := f.getFunctionPodStatus(namespace, 2*time.Minute)
	So(err, ShouldBeNil)

	host := fmt.Sprintf("http://%s.%s:8080", f.functionName, namespace)
	value, err := f.getFunctionOutput(host, 2*time.Minute)
	So(err, ShouldBeNil)
	So(value, ShouldContainSubstring, f.uuid)
}

func (f *functionTest) getFunctionOutput(host string, waitmax time.Duration) (string, error) {

	tick := time.Tick(2 * time.Second)
	timeout := time.After(waitmax)
	messages := ""

	for {
		select {
		case <-tick:
			resp, err := http.Post(host, "text/plain", bytes.NewBufferString(f.uuid))
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: kubelessV1.SchemeGroupVersion.String(),
		},
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
	watch, err := f.coreClient.CoreV1().Pods(namespace).Watch(metav1.ListOptions{LabelSelector: "function=" + f.functionName})
	if err != nil {
		return err
	}
	timeout := time.After(waitmax)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Pod did not start within given time  %v", waitmax)
		case event := <-watch.ResultChan():
			if event.Type == "ERROR" {
				return fmt.Errorf("%+v", event)
			}
			if event.Type == "MODIFIED" {
				pod, ok := event.Object.(*v1.Pod)
				if !ok {
					return fmt.Errorf("%v", event)
				}
				if pod.Status.Phase == "Running" {
					return nil
				}
			}
		}
	}

}
