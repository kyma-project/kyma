package testkit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-multierror"
	gatewayApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned/typed/gateway.kyma-project.io/v1alpha2"
	"github.com/pkg/errors"
	model "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsClient "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreClient "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	testServiceName         = "counter-service"
	testServicePort         = 8090
	testServiceImage        = "maladie/counterservice:latest"
	labelKey                = "component"
	healthEndpointFormat    = "https://counter-service.%s/health"
	endpointFormat          = "https://counter-service.%s"
	inClusterEndpointFormat = "http://counter-service.%s.svc.cluster.local:8090"
)

type TestService struct {
	apis        gatewayClient.ApiInterface
	deployments appsClient.DeploymentInterface
	services    coreClient.ServiceInterface
	HttpClient  *http.Client
	domain      string
	namespace   string
}

func NewTestService(httpClient *http.Client, deployments appsClient.DeploymentInterface, services coreClient.ServiceInterface, apis gatewayClient.ApiInterface, domain, namespace string) *TestService {
	return &TestService{
		HttpClient:  httpClient,
		domain:      domain,
		apis:        apis,
		deployments: deployments,
		services:    services,
		namespace:   namespace,
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
		return errors.Errorf("counter different then expected value: %v", count)
	}
	return nil
}

func (ts *TestService) DeleteTestService() error {
	errDeployment := ts.deployments.Delete(testServiceName, &v1.DeleteOptions{})
	errService := ts.services.Delete(testServiceName, &v1.DeleteOptions{})
	errApi := ts.apis.Delete(testServiceName, &v1.DeleteOptions{})
	err := multierror.Append(errDeployment, errService, errApi)
	return err.ErrorOrNil()
}

func (ts *TestService) GetTestServiceURL() string {
	return fmt.Sprintf(endpointFormat, ts.domain)
}

func (ts *TestService) GetInClusterTestServiceURL() string {
	return fmt.Sprintf(inClusterEndpointFormat, ts.namespace)
}

func (ts *TestService) getHealthEndpointURL() string {
	return fmt.Sprintf(healthEndpointFormat, ts.domain)
}

func (ts *TestService) createDeployment() error {
	rs := int32(1)
	deployment := &model.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name: testServiceName,
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
	_, err := ts.deployments.Create(deployment)
	return err
}

func (ts *TestService) createService() error {
	service := &core.Service{

		ObjectMeta: v1.ObjectMeta{
			Name: testServiceName,
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
	_, err := ts.services.Create(service)
	return err
}

func (ts *TestService) createAPI() error {
	api := &gatewayApi.Api{
		ObjectMeta: v1.ObjectMeta{
			Name: testServiceName,
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
	_, err := ts.apis.Create(api)
	return err
}
