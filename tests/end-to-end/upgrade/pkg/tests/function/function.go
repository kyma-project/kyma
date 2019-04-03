package function

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"strings"
	"os"
	"bytes"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	restclient "k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/api/resource"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	instr "k8s.io/apimachinery/pkg/util/intstr"
)

type FunctionUpgradeTest struct {
	functionName, uuid string
	kubelessClient     *kubeless.Clientset
	coreClient         *kubernetes.Clientset
	apiClient          *kyma.Clientset
	nSpace             string
	hostName           string
}

func int32Ptr(i int32) *int32 { return &i }

func NewFunctionUpgradeTest(config *restclient.Config) (*FunctionUpgradeTest) {

	kubelessClient, err := kubeless.NewForConfig(config)
	if err != nil {
		return &FunctionUpgradeTest{}
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &FunctionUpgradeTest{}
	}

	apiClient, err := kyma.NewForConfig(config)
	if err != nil {
		return &FunctionUpgradeTest{}
	}

	domainName := os.Getenv("DOMAIN")
	if len(domainName) == 0 {
		domainName = "kyma.local"
	}

	nSpace := strings.ToLower("FunctionUpgradeTest")
	functionName := "hello"
	hostName := fmt.Sprintf("%s-%s.%s", functionName, nSpace, domainName)
	return &FunctionUpgradeTest{
		kubelessClient: kubelessClient,
		coreClient:     coreClient,
		functionName:   functionName,
		uuid:           uuid.New().String(),
		nSpace:         nSpace,
		hostName:       hostName,
		apiClient:      apiClient,
	}
}

func (f *FunctionUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("FunctionUpgradeTest creating resources")
	f.nSpace = namespace

	err := f.createFunction()
	if err != nil {
		return err
	}

	err = f.createApi()
	if err != nil {
		log.Printf("create api %v", err)
		return err
	}

	return nil
}

func (f *FunctionUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("FunctionUpgradeTest testing resources")
	err := f.getFunctionPodStatus(10 * time.Minute)
	if err != nil {
		return err
	}

	host := fmt.Sprintf("https://%s", f.hostName)

	value, err := f.getFunctionOutput(host, 2*time.Minute, log)
	if err != nil {
		return err
	}

	if !strings.Contains(value, f.uuid) {
		return fmt.Errorf("could not get expected function output:\n %v\n output:\n %v", f.uuid, value)
	}

	return nil
}

func (f *FunctionUpgradeTest) getFunctionOutput(host string, waitmax time.Duration, log logrus.FieldLogger) (string, error) {
	log.Println("FunctionUpgradeTest function output")
	log.Printf("\nHost: %s", host)

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

func (f *FunctionUpgradeTest) createFunction() (error) {
	resources := make(map[corev1.ResourceName]resource.Quantity)
	resources[corev1.ResourceCPU] = resource.MustParse("100m")
	resources[corev1.ResourceMemory] = resource.MustParse("128Mi")

	annotations := make(map[string]string)
	annotations["function-size"] = "S"

	podContainers := []corev1.Container{}
	podContainer := corev1.Container{
		Name: f.functionName,
		Resources: corev1.ResourceRequirements{
			Limits:   resources,
			Requests: resources,
		},
	}

	podContainers = append(podContainers, podContainer)

	functionDeployment := extensionsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.functionName,
			Labels: map[string]string{
				"function": f.functionName,
			},
		},
		Spec: extensionsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: podContainers,
				},
			},
		},
	}

	serviceSelector := make(map[string]string)
	serviceSelector["created-by"] = "kubeless"
	serviceSelector["function"] = f.functionName

	function := &kubelessV1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:        f.functionName,
			Annotations: annotations,
		},
		Spec: kubelessV1.FunctionSpec{
			Handler:             "handler.main",
			Runtime:             "nodejs8",
			Function:            `module.exports = { main: function(event, context) { return(event.data) } }`,
			FunctionContentType: "text",
			ServiceSpec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "http-function-port",
						Port:       8080,
						Protocol:   corev1.ProtocolTCP,
						TargetPort: instr.FromInt(8080),
					},
				},
				Selector: serviceSelector,
			},
			Deployment: functionDeployment,
		},
	}
	_, err := f.kubelessClient.KubelessV1beta1().Functions(f.nSpace).Create(function)
	return err
}

func (f *FunctionUpgradeTest) getFunctionPodStatus(waitmax time.Duration) error {

	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			pods, err := f.coreClient.CoreV1().Pods(f.nSpace).List(metav1.ListOptions{LabelSelector: "function=" + f.functionName})
			if err != nil {
				return err
			}
			return fmt.Errorf("pod did not start within given time  %v: %+v", waitmax, pods)
		case <-tick:
			pods, err := f.coreClient.CoreV1().Pods(f.nSpace).List(metav1.ListOptions{LabelSelector: "function=" + f.functionName})
			if err != nil {
				return err
			}
			if len(pods.Items) == 0 {
				break
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

func (f *FunctionUpgradeTest) createApi() (error) {
	authEnabled := false
	servicePort := 8080

	api := &kymaApi.Api{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.functionName,
		},
		Spec: kymaApi.ApiSpec{
			AuthenticationEnabled: &authEnabled,
			Authentication:        []kymaApi.AuthenticationRule{},
			Hostname:              f.hostName,
			Service: kymaApi.Service{
				Name: f.functionName,
				Port: servicePort,
			},
		},
	}

	_, err := f.apiClient.GatewayV1alpha2().Apis(f.nSpace).Create(api)
	return err
}
