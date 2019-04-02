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
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	instr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type apiControllerTest struct {
	functionName   string
	uuid           string
	domainName     string
	hostName       string
	kubelessClient *kubeless.Clientset
	coreClient     *kubernetes.Clientset
	apiClient      *kyma.Clientset
}

func NewApiControllerTest() (apiControllerTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return apiControllerTest{}, err
	}

	kubelessClient, err := kubeless.NewForConfig(restConfig)
	if err != nil {
		return apiControllerTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return apiControllerTest{}, err
	}

	apiClient, err := kyma.NewForConfig(restConfig)
	if err != nil {
		return apiControllerTest{}, err
	}

	functionName := "apicontroller"
	domainName := os.Getenv("DOMAIN")

	return apiControllerTest{
		kubelessClient: kubelessClient,
		coreClient:     coreClient,
		apiClient:      apiClient,
		functionName:   functionName,
		domainName:     domainName,
		hostName:       functionName + "." + domainName,
		uuid:           uuid.New().String(),
	}, nil
}

func (t apiControllerTest) CreateResources(namespace string) {
	_, err := t.createFunction(namespace)
	So(err, ShouldBeNil)

	_, err = t.createApi(namespace)
	So(err, ShouldBeNil)
}

func (t apiControllerTest) TestResources(namespace string) {
	err := t.getFunctionPodStatus(namespace, 2*time.Minute)
	So(err, ShouldBeNil)

	err = t.callFunctionWithoutToken(2 * time.Minute)
	So(err, ShouldBeNil)

	token, err := fetchDexToken()
	So(err, ShouldBeNil)

	err = t.callFunctionWithToken(token, 2*time.Minute)
	So(err, ShouldBeNil)
}

func (t apiControllerTest) DeleteResources(namespace string) {
	// There is not need to be implemented for this test.
}

func (t apiControllerTest) callFunctionWithoutToken(waitmax time.Duration) error {
	tick := time.Tick(2 * time.Second)
	timeout := time.After(waitmax)
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

func (t apiControllerTest) callFunctionWithToken(token string, waitmax time.Duration) error {
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

func (t apiControllerTest) createApi(namespace string) (*apiv1alpha2.Api, error) {
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
						JwksUri: "http://dex-service.kyma-system.svc.cluster.local:5556/keys",
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

	return t.apiClient.GatewayV1alpha2().Apis(namespace).Create(api)
}

func (t apiControllerTest) createFunction(namespace string) (*kubelessV1.Function, error) {
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
