package testkit

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-multierror"
	gatewayApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned/typed/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/resourceskit"
	"github.com/pkg/errors"
	"io/ioutil"
	model "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
)

const (
	testServiceName            = "counter-service"
	testServicePort            = 8090
	testServiceImage           = "maladie/counterservice:latest"
	labelKey                   = "component"
	healthEndpointFormat       = "http://%s.%s:%v/health"
	healthEndpointFormatLocal  = "https://counter-service.%s/health"
	counterEndpointFormat      = "http://%s.%s:%v/counter"
	endpointFormatLocal = "https://counter-service.%s"
)

type TestService struct {
	K8sResourcesClient resourceskit.K8sResourcesClient
	apis               gatewayClient.ApisGetter
	HttpClient         *http.Client
	domain             string
}

func NewTestService(k8sClient resourceskit.K8sResourcesClient, httpClient *http.Client, apis gatewayClient.ApisGetter, domain string) *TestService {
	return &TestService{
		K8sResourcesClient: k8sClient,
		HttpClient:         httpClient,
		domain:             domain,
		apis:               apis,
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
		return errors.Errorf("counter different then expected value: %v", count)
	}
	return nil
}

func (ts *TestService) DeleteTestService() error {
	errDeployment := ts.K8sResourcesClient.DeleteDeployment(testServiceName, &v1.DeleteOptions{})
	errService := ts.K8sResourcesClient.DeleteService(testServiceName, &v1.DeleteOptions{})
	errApi := ts.apis.Apis(ts.K8sResourcesClient.GetNamespace()).Delete(testServiceName, &v1.DeleteOptions{})
	err := multierror.Append(errDeployment, errService, errApi)
	return err.ErrorOrNil()
}

func (ts *TestService) GetTestServiceURL() string {
	return fmt.Sprintf(endpointFormatLocal, ts.domain)
}

func (ts *TestService) getHealthEndpointURL() string {
	return fmt.Sprintf(healthEndpointFormatLocal, ts.domain)
}

func (ts *TestService) createDeployment() error {
	rs := int32(1)
	deployment := &model.Deployment{
		TypeMeta: v1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      testServiceName,
			Namespace: ts.K8sResourcesClient.GetNamespace(),
			Labels: map[string]string{
				labelKey: testServiceName,
			},
		},
		Spec: model.DeploymentSpec{
			Replicas: &rs,
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					labelKey: testServiceName,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						labelKey: testServiceName,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  testServiceName,
							Image: testServiceImage,
							Ports: []core.ContainerPort{
								{ContainerPort: testServicePort},
							},
						},
					},
				},
			},
		},
	}
	_, err := ts.K8sResourcesClient.CreateDeployment(deployment)
	return err
}

func (ts *TestService) createService() error {
	service := &core.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      testServiceName,
			Namespace: ts.K8sResourcesClient.GetNamespace(),
		},
		Spec: core.ServiceSpec{
			Type: "ClusterIP",
			Ports: []core.ServicePort{
				{
					Port:       testServicePort,
					TargetPort: intstr.FromInt(testServicePort),
				},
			},
			Selector: map[string]string{
				labelKey: testServiceName,
			},
		},
	}
	_, err := ts.K8sResourcesClient.CreateService(service)
	return err
}

func (ts *TestService) createAPI() error {
	api := &gatewayApi.Api{
		ObjectMeta: v1.ObjectMeta{
			Name:      testServiceName,
			Namespace: ts.K8sResourcesClient.GetNamespace(),
		},
		Spec: gatewayApi.ApiSpec{
			Service: gatewayApi.Service{
				Name: testServiceName,
				Port: testServicePort,
			},
			Hostname:       fmt.Sprintf("%s.%s", testServiceName, ts.domain),
			Authentication: []gatewayApi.AuthenticationRule{},
		},
	}
	_, err := ts.apis.Apis(ts.K8sResourcesClient.GetNamespace()).Create(api)
	return err
}
