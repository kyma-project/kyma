package backupe2e

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	dex "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/fetch-dex-token"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/google/uuid"
	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	apiv1alpha2 "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	gateway "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	instr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ApiControllerTest struct {
	functionName      string
	uuid              string
	domainName        string
	hostName          string
	kubelessInterface kubeless.Interface
	coreInterface     kubernetes.Interface
	apiInterface      gateway.Interface
}

func NewApiControllerTestFromEnv() (ApiControllerTest, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return ApiControllerTest{}, err
	}

	kubelessClient, err := kubeless.NewForConfig(config)
	if err != nil {
		return ApiControllerTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return ApiControllerTest{}, err
	}

	gatewayClient, err := gateway.NewForConfig(config)
	if err != nil {
		return ApiControllerTest{}, err
	}

	domainName := os.Getenv("DOMAIN")

	return NewApiControllerTest(gatewayClient, coreClient, kubelessClient, domainName), nil
}

func NewApiControllerTest(gatewayInterface gateway.Interface, coreInterface kubernetes.Interface, kubelessInterface kubeless.Interface, domainName string) ApiControllerTest {
	functionName := "apicontroller"
	return ApiControllerTest{
		kubelessInterface: kubelessInterface,
		coreInterface:     coreInterface,
		apiInterface:      gatewayInterface,
		functionName:      functionName,
		domainName:        domainName,
		hostName:          functionName + "." + domainName,
		uuid:              uuid.New().String(),
	}
}

func (t ApiControllerTest) CreateResources(namespace string) {
	err := t.CreateResourcesError(namespace)
	So(err, ShouldBeNil)
}

func (t ApiControllerTest) CreateResourcesError(namespace string) error {
	_, err := t.createFunction(namespace)
	if err != nil {
		return err
	}

	_, err = t.createApi(namespace)
	if err != nil {
		return err
	}

	return nil
}

func (t ApiControllerTest) TestResources(namespace string) {
	err := t.TestResourcesError(namespace)
	So(err, ShouldBeNil)
}

func (t ApiControllerTest) TestResourcesError(namespace string) error {
	err := t.getFunctionPodStatus(namespace, 2*time.Minute)
	if err != nil {
		return err
	}

	err = t.callFunctionWithoutToken(2 * time.Minute)
	if err != nil {
		return err
	}

	token, err := fetchDexToken()
	if err != nil {
		return err
	}

	err = t.callFunctionWithToken(token, 2*time.Minute)
	if err != nil {
		return err
	}

	return nil
}

func (t ApiControllerTest) callFunctionWithoutToken(waitMax time.Duration) error {

	tick := time.Tick(2 * time.Second)
	timeout := time.After(waitMax)
	messages := ""
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	for {
		select {
		case <-tick:
			host := fmt.Sprintf("https://%s", t.hostName)
			resp, err := http.Get(host)
			if err != nil {
				messages += fmt.Sprintf("%+v\n", err)
				break
			}
			if resp.StatusCode == http.StatusUnauthorized {
				return nil
			}
			messages += fmt.Sprintf("%+v", err)

		case <-timeout:
			return fmt.Errorf("Could not get function output:\n %v", messages)
		}
	}
}

func (t ApiControllerTest) callFunctionWithToken(token string, waitmax time.Duration) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	host := fmt.Sprintf("https://%s", t.hostName)
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	tick := time.Tick(2 * time.Second)
	timeout := time.After(waitmax)
	messages := ""

	for {
		select {
		case <-tick:

			resp, err := client.Do(req)
			if err != nil {
				messages += fmt.Sprintf("%+v\n", err)
				break
			}
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			messages += fmt.Sprintf("%+v", err)

		case <-timeout:
			return fmt.Errorf("Could not get function output:\n %v", messages)
		}
	}
}

func (t ApiControllerTest) createApi(namespace string) (*apiv1alpha2.Api, error) {
	authEnabled := true
	servicePort := 8080

	api := &apiv1alpha2.Api{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.functionName,
		},
		Spec: apiv1alpha2.ApiSpec{
			AuthenticationEnabled: &authEnabled,
			Authentication: []apiv1alpha2.AuthenticationRule{
				{
					Type: apiv1alpha2.JwtType,
					Jwt: apiv1alpha2.JwtAuthentication{
						Issuer:  "https://dex." + t.domainName,
						JwksUri: "http://dex-service.gateway-system.svc.cluster.local:5556/keys",
					},
				},
			},
			Hostname: t.hostName,
			Service: apiv1alpha2.Service{
				Name: t.functionName,
				Port: servicePort,
			},
		},
	}

	return t.apiInterface.GatewayV1alpha2().Apis(namespace).Create(api)
}

func (t ApiControllerTest) createFunction(namespace string) (*kubelessV1.Function, error) {
	resources := make(map[corev1.ResourceName]resource.Quantity)
	resources[corev1.ResourceCPU] = resource.MustParse("100m")
	resources[corev1.ResourceMemory] = resource.MustParse("128Mi")

	annotations := make(map[string]string)
	annotations["function-size"] = "S"

	var podContainers []corev1.Container
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
	return t.kubelessInterface.KubelessV1beta1().Functions(namespace).Create(function)
}

func (t ApiControllerTest) getFunctionPodStatus(namespace string, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			pods, err := t.coreInterface.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + t.functionName})
			if err != nil {
				return err
			}
			return fmt.Errorf("Pod did not start within given time  %v: %+v", waitmax, pods)
		case <-tick:
			pods, err := t.coreInterface.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + t.functionName})
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
			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}
			if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
				return fmt.Errorf("Function in state %v: \n%+v", pod.Status.Phase, pod)
			}
		}
	}
}

func fetchDexToken() (string, error) {
	config, err := dex.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	token, err := dex.Authenticate(config.IdProviderConfig)
	if err != nil {
		log.Fatal(err)
	}
	return token, nil
}

func (t ApiControllerTest) DeleteResources() {
	// There is not need to be implemented for this test.
}