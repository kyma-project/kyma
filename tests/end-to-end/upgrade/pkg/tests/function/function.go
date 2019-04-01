package function

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"strings"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"

	restclient "k8s.io/client-go/rest"
)

type FunctionUpgradeTest struct {
	functionName, uuid string
	kubelessClient     *kubeless.Clientset
	coreClient         *kubernetes.Clientset
}

func NewFunctionUpgradeTest(config *restclient.Config) (*FunctionUpgradeTest) {

	kubelessClient, err := kubeless.NewForConfig(config)
	if err != nil {
		return &FunctionUpgradeTest{}
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &FunctionUpgradeTest{}
	}

	return &FunctionUpgradeTest{
		kubelessClient: kubelessClient,
		coreClient:     coreClient,
		functionName:   "hello",
		uuid:           uuid.New().String(),
	}
}

func (f *FunctionUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("FunctionUpgradeTest creating resources")
	_, err := f.createFunction(log, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (f *FunctionUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("FunctionUpgradeTest testing resources")
	err := f.getFunctionPodStatus(namespace, 10*time.Minute)
	if err != nil {
		return err
	}

	host := fmt.Sprintf("http://%s.%s:8080", f.functionName, namespace)

	value, err := f.getFunctionOutput(host, 2*time.Minute)
	if err != nil {
		return err
	}

	if !strings.Contains(value, f.uuid) {
		return fmt.Errorf("could not get expected function output:\n %v\n output:\n %v", f.uuid, value)
	}

	return nil
}

func (f *FunctionUpgradeTest) getFunctionOutput(host string, waitmax time.Duration) (string, error) {

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
			return "", fmt.Errorf("could not get function output:\n %v", messages)
		}
	}

}

func (f *FunctionUpgradeTest) createFunction(log logrus.FieldLogger, namespace string) (*kubelessV1.Function, error) {
	log.Println("FunctionUpgradeTest creating function")
	function := &kubelessV1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.functionName,
			Labels: map[string]string{
				"function": f.functionName,
			},
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
	_, err := f.kubelessClient.KubelessV1beta1().Functions(namespace).Create(function)
	if err != nil {
		return nil, err
	}

	return function, nil
}

func (f *FunctionUpgradeTest) getFunctionPodStatus(namespace string, waitmax time.Duration) error {

	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			pods, err := f.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + f.functionName})
			if err != nil {
				return err
			}
			return fmt.Errorf("pod did not start within given time  %v: %+v", waitmax, pods)
		case <-tick:
			pods, err := f.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + f.functionName})
			if err != nil {
				return err
			}
			if len(pods.Items) == 0 {
				break
			}

			if len(pods.Items) > 1 {
				return fmt.Errorf("deployed 1 pod, got %v: %+v", len(pods.Items), pods)
			}

			pod := pods.Items[0]
			// If Pod condition is not ready the for will continue until timeout
			if len(pod.Status.Conditions) > 0 {
				conditions := pod.Status.Conditions
				for _, cond := range conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						return nil
					}
				}
			}

			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
				return fmt.Errorf("function in state %v: \n%+v", pod.Status.Phase, pod)
			}
		}
	}
}