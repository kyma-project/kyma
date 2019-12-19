package api_gateway

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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	dex "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/fetch-dex-token"
	"k8s.io/apimachinery/pkg/api/resource"

	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	apiRulev1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	instr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/dynamic"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
)

var (
	functionRes    = schema.GroupVersionResource{Version: kubelessv1beta1.SchemeGroupVersion.Version, Group: kubelessv1beta1.SchemeGroupVersion.Group, Resource: "functions"}
	hydraClientRes = schema.GroupVersionResource{Group: "hydra.ory.sh", Version: "v1alpha1", Resource: "oauth2clients"}
	secretRes      = schema.GroupVersionResource{Group: corev1.GroupName, Version: "v1", Resource: "secrets"}
	apiRuleRes     = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
)

type ApiGatewayTest struct {
	functionName  string
	secretName    string
	domainName    string
	hostName      string
	coreInterface kubernetes.Interface
	dynInterface  dynamic.Interface
	idpConfig     dex.IdProviderConfig
}

func NewApiGatewayTest(coreInterface kubernetes.Interface, dynamicInterface dynamic.Interface, domainName string, dexConfig dex.IdProviderConfig) ApiGatewayTest {
	functionName := "apigateway"
	secretName := "api-gateway-upgrade-tests"
	return ApiGatewayTest{
		coreInterface: coreInterface,
		dynInterface:  dynamicInterface,
		functionName:  functionName,
		secretName:    secretName,
		domainName:    domainName,
		hostName:      functionName + "." + domainName,
		idpConfig:     dexConfig,
	}
}

func (t ApiGatewayTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.CreateResourcesError(namespace)
}

func (t ApiGatewayTest) CreateResourcesError(namespace string) error {
	_, err := t.createFunction(namespace)
	if err != nil {
		return err
	}

	_, err = t.createHydraClientSecret(namespace)
	if err != nil {
		return err
	}

	_, err = t.createHydraClient(namespace)
	if err != nil {
		return err
	}

	_, err = t.createApiRule(namespace)
	if err != nil {
		return err
	}

	return nil
}

func (t ApiGatewayTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.TestResourcesError(namespace)
}

func (t ApiGatewayTest) TestResourcesError(namespace string) error {
	err := t.getFunctionPodStatus(namespace, 2*time.Minute)
	if err != nil {
		return err
	}

	err = t.callFunctionWithoutToken()
	if err != nil {
		return err
	}

	jwtToken, err := t.fetchDexToken()
	if err != nil {
		return err
	}

	err = t.callFunctionWithJWTToken(jwtToken)
	if err != nil {
		return err
	}

	//TODO fetch oauth token and call lambda

	return nil
}

func (t ApiGatewayTest) callFunctionWithoutToken() error {
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

func (t ApiGatewayTest) callFunctionWithJWTToken(token string) error {
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

func (t ApiGatewayTest) createApiRule(namespace string) (*unstructured.Unstructured, error) {

	fmt.Println("createApiRule")

	var gateway = "kyma-gateway.kyma-system.svc.cluster.local"
	var servicePort uint32 = 8080

	jwtConfigJSON := fmt.Sprintf(`{"trusted_issuers": ["https://dex.%s"]}`, t.domainName)

	oauthConfigJSON := `{"required_scope": ["read"]}`

	strategies := []*rulev1alpha1.Authenticator{
		{
			Handler: &rulev1alpha1.Handler{
				Name: "jwt",
				Config: &runtime.RawExtension{
					Raw: []byte(jwtConfigJSON),
				},
			},
		},
		{
			Handler: &rulev1alpha1.Handler{
				Name: "oauth2_introspection",
				Config: &runtime.RawExtension{
					Raw: []byte(oauthConfigJSON),
				},
			},
		},
	}

	apiRule := &apiRulev1alpha1.APIRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIRule",
			APIVersion: apiRulev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.functionName,
		},
		Spec: apiRulev1alpha1.APIRuleSpec{
			Gateway: &gateway,
			Service: &apiRulev1alpha1.Service{
				Name: &t.functionName,
				Port: &servicePort,
				Host: &t.hostName,
			},
			Rules: []apiRulev1alpha1.Rule{
				{
					Path:             "/.*",
					Methods:          []string{"GET"},
					AccessStrategies: strategies,
				},
			},
		},
	}

	unstructured, err := toUnstructured(&apiRule)
	if err != nil {
		return nil, err
	}

	return t.dynInterface.Resource(apiRuleRes).Namespace(namespace).Create(unstructured, metav1.CreateOptions{})
}

func (t ApiGatewayTest) createFunction(namespace string) (*unstructured.Unstructured, error) {

	fmt.Println("createFunction")

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

	functionDeployment := appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.functionName,
		},
		Spec: appv1.DeploymentSpec{
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

	function := &kubelessv1beta1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: kubelessv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        t.functionName,
			Annotations: annotations,
		},
		Spec: kubelessv1beta1.FunctionSpec{
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

	unstructured, err := toUnstructured(&function)
	if err != nil {
		return nil, err
	}

	return t.dynInterface.Resource(functionRes).Namespace(namespace).Create(unstructured, metav1.CreateOptions{})
}

func (t ApiGatewayTest) createHydraClientSecret(namespace string) (*unstructured.Unstructured, error) {

	fmt.Println("createHydraClientSecret")

	secretData := make(map[string]string)
	secretData["client_id"] = "dummy_client"
	secretData["client_secret"] = "dummy_secret"

	hydraClientSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.secretName,
		},
		StringData: secretData,
	}

	unstructured, err := toUnstructured(&hydraClientSecret)
	if err != nil {
		return nil, err
	}

	return t.dynInterface.Resource(secretRes).Namespace(namespace).Create(unstructured, metav1.CreateOptions{})
}

func (t ApiGatewayTest) createHydraClient(namespace string) (*unstructured.Unstructured, error) {

	fmt.Println("createHydraClient")

	hydraClient := &hydrav1alpha1.OAuth2Client{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OAuth2Client",
			APIVersion: hydrav1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.functionName,
		},
		Spec: hydrav1alpha1.OAuth2ClientSpec{
			GrantTypes: []hydrav1alpha1.GrantType{"client_credentials"},
			Scope:      "read",
			SecretName: "api-gateway-upgrade-tests",
		},
	}

	unstructured, err := toUnstructured(&hydraClient)
	if err != nil {
		return nil, err
	}

	return t.dynInterface.Resource(hydraClientRes).Namespace(namespace).Create(unstructured, metav1.CreateOptions{})
}

func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&obj)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: object}, nil
}

func (t ApiGatewayTest) getFunctionPodStatus(namespace string, waitmax time.Duration) error {
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

func (t ApiGatewayTest) fetchDexToken() (string, error) {
	return dex.Authenticate(t.idpConfig)
}

func (t ApiGatewayTest) int32Ptr(i int32) *int32 { return &i }
