package backupe2e

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/resource"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	apiv1alpha2 "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	instr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/smartystreets/goconvey/convey"

	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
)

type apiControllerTest struct {
	functionName   string
	uuid           string
	kubelessClient *kubeless.Clientset
	coreClient     *kubernetes.Clientset
	apiClient      *kyma.Clientset
}

func NewApiControllerTest() (apiControllerTest, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return apiControllerTest{}, err
	}

	kubelessClient, err := kubeless.NewForConfig(config)
	if err != nil {
		return apiControllerTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return apiControllerTest{}, err
	}

	apiClient, err := kyma.NewForConfig(config)
	if err != nil {
		return apiControllerTest{}, err
	}

	return apiControllerTest{
		kubelessClient: kubelessClient,
		coreClient:     coreClient,
		apiClient:      apiClient,
		functionName:   "apicontroller",
		uuid:           uuid.New().String(),
	}, nil
}

func (t apiControllerTest) CreateResources(namespace string) {
	_, err := t.createFunction(namespace)
	if err != nil {
		log.Println("%+v", err)
	}

	_, err = t.createApi(namespace)
	if err != nil {
		log.Println("%+v", err)
	}
	//So(err, ShouldBeNil)
}

func (t apiControllerTest) TestResources(namespace string) {
	err := t.getFunctionPodStatus(namespace, 2*time.Minute)
	So(err, ShouldBeNil)

	host := fmt.Sprintf("http://%s.%s:8080", t.functionName, namespace)
	value, err := t.getFunctionOutput(host, 2*time.Minute)
	So(err, ShouldBeNil)
	So(value, ShouldContainSubstring, t.uuid)
}

func (t apiControllerTest) getFunctionOutput(host string, waitmax time.Duration) (string, error) {

	tick := time.Tick(2 * time.Second)
	timeout := time.After(waitmax)
	messages := ""

	for {
		select {
		case <-tick:
			resp, err := http.Post(host, "text/plain", bytes.NewBufferString(t.uuid))
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

func (t apiControllerTest) createApi(namespace string) (*apiv1alpha2.Api, error) {
	authRules := []apiv1alpha2.AuthenticationRule{}
	authRule := apiv1alpha2.AuthenticationRule{
		Type: apiv1alpha2.JwtType,
		Jwt: apiv1alpha2.JwtAuthentication{
			Issuer:  "https://dex.kyma.local",
			JwksUri: "http://dex-service.kyma-system.svc.cluster.local:5556/keys",
		},
	}

	authRules = append(authRules, authRule)

	api := &apiv1alpha2.Api{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.functionName,
		},
		Spec: apiv1alpha2.ApiSpec{
			Authentication: authRules,
			Hostname:       "apicontroller-stage.kyma.local",
			Service: apiv1alpha2.Service{
				Name: t.functionName,
				Port: 8080,
			},
		},
	}

	return t.apiClient.GatewayV1alpha2().Apis(namespace).Create(api)
}

func (t apiControllerTest) createFunction(namespace string) (*kubelessV1.Function, error) {
	functionServicePort := corev1.ServicePort{
		Name:       "http-function-port",
		Port:       8080,
		Protocol:   corev1.ProtocolTCP,
		TargetPort: instr.FromInt(8080),
	}

	functionServicePorts := []corev1.ServicePort{}

	functionServicePorts = append(functionServicePorts, functionServicePort)

	var repl = int32Ptr(1)

	resources := make(map[corev1.ResourceName]resource.Quantity)
	resources[corev1.ResourceCPU] = resource.MustParse("100m")
	resources[corev1.ResourceMemory] = resource.MustParse("128Mi")

	annotations := make(map[string]string)
	annotations["function-size"] = "S"

	podContainers := []corev1.Container{}
	podContainer := corev1.Container{
		Name: t.functionName,
		Resources: corev1.ResourceRequirements{
			Limits:   resources,
			Requests: resources,
		},
	}

	podContainers = append(podContainers, podContainer)

	functionDeployment := extensionsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.functionName,
		},
		Spec: extensionsv1.DeploymentSpec{
			Replicas: repl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: podContainers,
				},
			},
		},
	}

	serviceSelector := make(map[string]string)
	serviceSelector["created-by"] = "kubeless"
	serviceSelector["function"] = t.functionName

	function := &kubelessV1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:        t.functionName,
			Annotations: annotations,
		},
		Spec: kubelessV1.FunctionSpec{
			Handler: "handler.authorize",
			Runtime: "nodejs8",
			Function: `module.exports = {
				authorize: function(event, context) {
				  return(event.data)
				}
			  }`,
			FunctionContentType: "text",
			ServiceSpec: corev1.ServiceSpec{
				Ports:    functionServicePorts,
				Selector: serviceSelector,
			},
			Deployment: functionDeployment,
		},
	}
	return t.kubelessClient.KubelessV1beta1().Functions(namespace).Create(function)
}

func (t apiControllerTest) getFunctionPodStatus(namespace string, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			pods, err := t.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + t.functionName})
			if err != nil {
				return err
			}
			return fmt.Errorf("Pod did not start within given time  %v: %+v", waitmax, pods)
		case <-tick:
			pods, err := t.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + t.functionName})
			if err != nil {
				return err
			}
			if len(pods.Items) == 0 {
				break
			}

			if len(pods.Items) > 1 {
				return fmt.Errorf("Deployed 1 pod, got %v: %+v", len(pods.Items), pods)
			}

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
