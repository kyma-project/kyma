package testkit

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	model "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
)

const (
	testServiceName       = "counter-service"
	testServicePort       = 8090
	testServiceImage      = "maladie/counterservice:latest"
	labelKey              = "component"
	healthEndpointFormat  = "http://%s.%s/health"
	counterEndpointFormat = "http://%s.%s/counter"
)

type TestService interface {
	CreateTestService() error
	DeleteTestService() error
	CheckValue() (int, error)
	CheckIfReady() (bool, error)
	GetTestServiceURL() string
}

type testService struct {
	K8sResourcesClient resourceskit.K8sResourcesClient
	HttpClient         http.Client
}

func NewTestService(k8sClient resourceskit.K8sResourcesClient, httpClient http.Client) TestService {
	return &testService{
		K8sResourcesClient: k8sClient,
		HttpClient:         httpClient,
	}
}

func (ts *testService) CreateTestService() error {
	var e error
	e = ts.createDeployment()
	if e != nil {
		return e
	}
	e = ts.createService()
	if e != nil {
		return e
	}
	return nil
}

func (ts *testService) CheckValue() (int, error) {

	url := ts.GetTestServiceURL()

	resp, err := ts.HttpClient.Get(url)

	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var response struct {
		counter int
	}

	e := json.NewDecoder(resp.Body).Decode(&response)

	if e != nil {
		return 0, e
	}

	return response.counter, nil
}

func (ts *testService) CheckIfReady() (bool, error) {

	url := ts.getHealthEndpointURL()

	resp, err := ts.HttpClient.Get(url)

	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

func (ts *testService) DeleteTestService() error {
	var e error

	e = ts.K8sResourcesClient.DeleteDeployment(testServiceName, &v1.DeleteOptions{})

	if e != nil {
		return e
	}

	e = ts.K8sResourcesClient.DeleteService(testServiceName, &v1.DeleteOptions{})

	if e != nil {
		return e
	}

	return nil
}

func (ts *testService) GetTestServiceURL() string {
	namespace := ts.K8sResourcesClient.GetNamespace()
	return fmt.Sprintf(counterEndpointFormat, testServiceName, namespace)
}

func (ts *testService) getHealthEndpointURL() string {
	namespace := ts.K8sResourcesClient.GetNamespace()
	return fmt.Sprintf(healthEndpointFormat, testServiceName, namespace)
}

func (ts *testService) createDeployment() error {
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
	_, e := ts.K8sResourcesClient.CreateDeployment(deployment)
	return e
}

func (ts *testService) createService() error {
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
	_, e := ts.K8sResourcesClient.CreateService(service)
	return e
}
