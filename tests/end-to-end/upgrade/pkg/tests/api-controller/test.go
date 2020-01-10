package apicontroller

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	dex "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/fetch-dex-token"
	"k8s.io/apimachinery/pkg/api/resource"

	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	apiv1alpha2 "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	gateway "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
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

// New creates new instance of Test
func NewAPIControllerTest(gatewayInterface gateway.Interface, coreInterface kubernetes.Interface, kubelessInterface kubeless.Interface, domainName string, dexConfig dex.IdProviderConfig) ApiControllerTest {
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

// CreateResources creates resources for tests
func (t ApiControllerTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
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

// TestResources tests if resources are working properly after upgrade
func (t ApiControllerTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	err := t.getFunctionPodStatus(namespace, 2*time.Minute)
	if err != nil {
		return err
	}

	err = t.callFunctionWithoutToken()
	if err != nil {
		return err
	}

	token, err := t.fetchDexToken()
	if err != nil {
		return err
	}

	err = t.callFunctionWithToken(token)
	if err != nil {
		return err
	}

	return nil
}

func (t ApiControllerTest) callFunctionWithoutToken() error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	err := retry.Do(func() error {
		host := fmt.Sprintf("https://%s", t.hostName)
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

func (t ApiControllerTest) callFunctionWithToken(token string) error {
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

	functionDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.functionName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: t.int32Ptr(1),
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
	const retriesCount = 10
	return retry.Do(
		func() error {
			pods, err := t.coreInterface.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "function=" + t.functionName})
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
			log.Printf("Function Pod Status exection #%d: %s\n", n, err)
		}),
	)
}

func (t ApiControllerTest) fetchDexToken() (string, error) {
	return dex.Authenticate(t.idpConfig)
}

func (t ApiControllerTest) int32Ptr(i int32) *int32 { return &i }
