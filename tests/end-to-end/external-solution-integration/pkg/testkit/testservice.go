package testkit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/dynamic"

	"github.com/hashicorp/go-multierror"
	apiRulev1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	testServiceNamePrefix   = "ctr-svc"
	testServicePort         = 8090
	testServiceImage        = "maladie/counterservice:latest"
	labelKey                = "component"
	healthEndpointFormat    = "https://%s.%s/health"
	endpointFormat          = "https://%s.%s"
	inClusterEndpointFormat = "http://%s.%s.svc.cluster.local:8090"
	gateway                 = "kyma-gateway.kyma-system.svc.cluster.local"
)

type TestService struct {
	apiRules        dynamic.ResourceInterface
	deployments     appsclient.DeploymentInterface
	services        coreclient.ServiceInterface
	HttpClient      *http.Client
	domain          string
	namespace       string
	testServiceName string
}

func NewTestService(httpClient *http.Client, deployments appsclient.DeploymentInterface, services coreclient.ServiceInterface, apiRules dynamic.ResourceInterface, domain, namespace string) *TestService {
	return &TestService{
		HttpClient:      httpClient,
		domain:          domain,
		apiRules:        apiRules,
		deployments:     deployments,
		services:        services,
		namespace:       namespace,
		testServiceName: fmt.Sprintf("%v-%v", testServiceNamePrefix, namespace),
	}
}

func (ts *TestService) CreateTestService() error {
	err := ts.createDeployment()
	if err != nil {
		return err
	}
	err = ts.createService()
	if err != nil {
		return err
	}
	return ts.createAPI()
}

func (ts *TestService) checkValue() (int, error) {

	url := ts.GetTestServiceURL()
	resp, err := ts.HttpClient.Get(url + "/counter")

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		return 0, errors.Errorf("error response: %s", body)
	}

	var response struct {
		Counter int `json:"counter"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return 0, err
	}

	return response.Counter, nil
}

func (ts *TestService) IsReady() error {

	url := ts.getHealthEndpointURL()
	resp, err := ts.HttpClient.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}
func (ts *TestService) WaitForCounterPodToUpdateValue(val int) error {
	count, err := ts.checkValue()
	if err != nil {
		return err
	}

	if count != val {
		return errors.Errorf("counter different then expected value: Got: %v but expected %v", count, val)
	}
	return nil
}

func (ts *TestService) DeleteTestService() error {
	errDeployment := ts.deployments.Delete(ts.testServiceName, &metav1.DeleteOptions{})
	errService := ts.services.Delete(ts.testServiceName, &metav1.DeleteOptions{})
	errApi := ts.apiRules.Delete(ts.testServiceName, &metav1.DeleteOptions{})
	err := multierror.Append(errDeployment, errService, errApi)
	return err.ErrorOrNil()
}

func (ts *TestService) GetTestServiceURL() string {
	return fmt.Sprintf(endpointFormat, ts.testServiceName, ts.domain)
}

func (ts *TestService) GetInClusterTestServiceURL() string {
	return fmt.Sprintf(inClusterEndpointFormat, ts.testServiceName, ts.namespace)
}

func (ts *TestService) getHealthEndpointURL() string {
	return fmt.Sprintf(healthEndpointFormat, ts.testServiceName, ts.domain)
}

func (ts *TestService) createDeployment() error {
	rs := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.testServiceName,
			Labels: map[string]string{
				labelKey: ts.testServiceName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &rs,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelKey: ts.testServiceName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						labelKey: ts.testServiceName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  ts.testServiceName,
							Image: testServiceImage,
							Ports: []v1.ContainerPort{
								{ContainerPort: testServicePort},
							},
						},
					},
				},
			},
		},
	}
	_, err := ts.deployments.Create(deployment)
	return err
}

func (ts *TestService) createService() error {
	service := &v1.Service{

		ObjectMeta: metav1.ObjectMeta{
			Name: ts.testServiceName,
		},
		Spec: v1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []v1.ServicePort{
				{
					Name:       "http-counter",
					Port:       testServicePort,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(testServicePort),
				},
			},
			Selector: map[string]string{
				labelKey: ts.testServiceName,
			},
		},
	}
	_, err := ts.services.Create(service)
	return err
}

func (ts *TestService) createAPI() error {
	port := uint32(testServicePort)
	host := fmt.Sprintf("%s.%s", ts.testServiceName, ts.domain)
	gateway := gateway
	apiRule := &apiRulev1alpha1.APIRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIRule",
			APIVersion: apiRulev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ts.testServiceName,
		},
		Spec: apiRulev1alpha1.APIRuleSpec{
			Service: &apiRulev1alpha1.Service{
				Name: &ts.testServiceName,
				Port: &port,
				Host: &host,
			},
			Gateway: &gateway,
			Rules: []apiRulev1alpha1.Rule{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					AccessStrategies: []*rulev1alpha1.Authenticator{
						{
							Handler: &rulev1alpha1.Handler{
								Name: "noop",
							},
						},
					},
				},
			},
		},
	}

	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&apiRule)
	if err != nil {
		return err
	}

	_, err = ts.apiRules.Create(&unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})
	return err
}
