package apicontroller

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/api/resource"

	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"

	apiv1alpha2 "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	gateway "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
	dex "github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/fetch-dex-token"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	instr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type ApiControllerTest struct {
	functionName      string
	domainName        string
	hostName          string
	kubelessInterface kubeless.Interface
	coreInterface     kubernetes.Interface
	apiInterface      gateway.Interface
	idpConfig         dex.IdProviderConfig
}

func NewApiControllerTestFromEnv() (ApiControllerTest, error) {

	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return ApiControllerTest{}, err
	}

	kubelessClient, err := kubeless.NewForConfig(restConfig)
	if err != nil {
		return ApiControllerTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return ApiControllerTest{}, err
	}

	gatewayClient, err := gateway.NewForConfig(restConfig)
	if err != nil {
		return ApiControllerTest{}, err
	}

	dexConfig, err := dex.LoadConfig()
	if err != nil {
		return ApiControllerTest{}, err
	}

	domainName := os.Getenv("DOMAIN")

	return NewApiControllerTest(gatewayClient, coreClient, kubelessClient, domainName, dexConfig.IdProviderConfig()), nil
}

func NewApiControllerTest(gatewayInterface gateway.Interface, coreInterface kubernetes.Interface, kubelessInterface kubeless.Interface, domainName string, dexConfig dex.IdProviderConfig) ApiControllerTest {
	functionName := "apicontroller"
	return ApiControllerTest{
		kubelessInterface: kubelessInterface,
		coreInterface:     coreInterface,
		apiInterface:      gatewayInterface,
		functionName:      functionName,
		domainName:        domainName,
		hostName:          functionName + "." + domainName,
		idpConfig:         dexConfig,
	}
}

func (act ApiControllerTest) CreateResources(t *testing.T, namespace string) {
	err := act.CreateResourcesError(t, namespace)
	require.NoError(t, err)
}

func (act ApiControllerTest) CreateResourcesError(t *testing.T, namespace string) error {
	_, err := act.createFunction(namespace)
	if err != nil {
		return err
	}

	_, err = act.createAPI(namespace)
	if err != nil {
		return err
	}

	return nil
}

func (act ApiControllerTest) TestResources(t *testing.T, namespace string) {
	err := act.TestResourcesError(t, namespace)
	require.NoError(t, err)
}

func (act ApiControllerTest) TestResourcesError(t *testing.T, namespace string) error {
	err := act.getFunctionPodStatus(t, namespace, 2*time.Minute)
	require.NoError(t, err)

	err = act.callFunctionWithoutToken()
	require.NoError(t, err)

	token, err := act.fetchDexToken()
	require.NoError(t, err)

	err = act.callFunctionWithToken(token)
	require.NoError(t, err)

	return nil
}

func (act ApiControllerTest) callFunctionWithoutToken() error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	err := retry.Do(func() error {
		host := fmt.Sprintf("https://%s", act.hostName)
		resp, err := http.Get(host)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return nil
		}
		rspBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Errorf("unexpected response %v: %s", resp.StatusCode, string(rspBody))
	})
	if err != nil {
		err = errors.Wrap(err, "cannot callFunctionWithoutToken")
	}
	return err
}

func (act ApiControllerTest) callFunctionWithToken(token string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	host := fmt.Sprintf("https://%s", act.hostName)
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	err = retry.Do(func() error {
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		rspBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Errorf("unexpected response %v: %s", resp.StatusCode, string(rspBody))
	})
	if err != nil {
		err = errors.Wrap(err, "cannot callFunctionWithToken")
	}
	return err
}

func (act ApiControllerTest) createAPI(namespace string) (*apiv1alpha2.Api, error) {
	authEnabled := true
	servicePort := 8080

	api := &apiv1alpha2.Api{
		ObjectMeta: metav1.ObjectMeta{
			Name: act.functionName,
		},
		Spec: apiv1alpha2.ApiSpec{
			AuthenticationEnabled: &authEnabled,
			Authentication: []apiv1alpha2.AuthenticationRule{
				{
					Type: apiv1alpha2.JwtType,
					Jwt: apiv1alpha2.JwtAuthentication{
						Issuer:  "https://dex." + act.domainName,
						JwksUri: "http://dex-service.kyma-system.svc.cluster.local:5556/keys",
					},
				},
			},
			Hostname: act.hostName,
			Service: apiv1alpha2.Service{
				Name: act.functionName,
				Port: servicePort,
			},
		},
	}

	return act.apiInterface.GatewayV1alpha2().Apis(namespace).Create(api)
}

func (act ApiControllerTest) createFunction(namespace string) (*kubelessV1.Function, error) {
	resources := make(map[corev1.ResourceName]resource.Quantity)
	resources[corev1.ResourceCPU] = resource.MustParse("100m")
	resources[corev1.ResourceMemory] = resource.MustParse("128Mi")

	annotations := make(map[string]string)
	annotations["function-size"] = "S"

	var podContainers []corev1.Container
	podContainer := corev1.Container{
		Name: act.functionName,
		Resources: corev1.ResourceRequirements{
			Limits:   resources,
			Requests: resources,
		},
	}

	podContainers = append(podContainers, podContainer)

	functionDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: act.functionName,
		},
		Spec: appsv1.DeploymentSpec{
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
	serviceSelector["function"] = act.functionName

	function := &kubelessV1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:        act.functionName,
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
	return act.kubelessInterface.KubelessV1beta1().Functions(namespace).Create(function)
}

func (act ApiControllerTest) getFunctionPodStatus(t *testing.T, namespace string, waitmax time.Duration) error {
	const retriesCount = 10
	return retry.Do(
		func() error {
			pods, err := act.coreInterface.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + act.functionName})
			if err != nil {
				return err
			}
			if len(pods.Items) == 0 {
				return errors.New("pod not scheduled yet")
			}
			if len(pods.Items) > 1 {
				jsonPods, err := json.Marshal(pods)
				if err != nil {
					return err
				}
				return errors.Errorf("expected 1 pod, got %d: %s", len(pods.Items), string(jsonPods))
			}

			pod := pods.Items[0]
			if pod.Status.Phase == corev1.PodRunning {
				return nil
			}
			return errors.Errorf("Function in state %v: \n%+v", pod.Status.Phase, pod)
		},
		retry.Attempts(retriesCount),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(waitmax/retriesCount), //doesn't have to be very precise
		retry.OnRetry(func(n uint, err error) {
			t.Logf("Function Pod Status exection #%d: %s\n", n, err)
		}),
	)
}

func (act ApiControllerTest) fetchDexToken() (string, error) {
	return dex.Authenticate(act.idpConfig)
}

func int32Ptr(i int32) *int32 { return &i }
