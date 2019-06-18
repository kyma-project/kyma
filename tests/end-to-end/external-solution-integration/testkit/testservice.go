package testkit

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	log "github.com/sirupsen/logrus"
	model "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	testServiceName       = "counter-service"
	testServicePort       = 8090
	testServiceImage      = "maladie/counterservice:latest"
	labelKey              = "component"
	healthEndpointFormat  = "http://%s.%s:%v/health"
	counterEndpointFormat = "http://%s.%s:%v"
)

type TestService interface {
	CreateTestService() error
	DeleteTestService() error
	checkValue() (int, error)
	IsReady() (bool, error)
	GetTestServiceURL() string
	WaitForCounterPodToUpdateValue() (bool, error)
}

type testService struct {
	K8sResourcesClient resourceskit.K8sResourcesClient
	HttpClient         http.Client
}

func NewTestService(k8sClient resourceskit.K8sResourcesClient) TestService {

	httpClient := newHttpClient(true)

	return &testService{
		K8sResourcesClient: k8sClient,
		HttpClient:         *httpClient,
	}
}

func (ts *testService) CreateTestService() error {
	err := ts.createDeployment()
	if err != nil {
		return err
	}
	err = ts.createService()
	if err != nil {
		return err
	}
	return nil
}

func (ts *testService) checkValue() (int, error) {

	url := ts.GetTestServiceURL() + "/counter"

	resp, err := ts.HttpClient.Get(url)

	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var response struct {
		Counter int `json:"counter"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return 0, err
	}

	return response.Counter, nil
}

func (ts *testService) IsReady() (bool, error) {

	url := ts.getHealthEndpointURL()
	log.WithFields(log.Fields{"url": url}).Debug("IsReady?")

	resp, _ := ts.HttpClient.Get(url)

	// lets ignore all errors here, shall we?

	// if err != nil {
	// return false, err
	// }
	if resp != nil {
		if resp.StatusCode == http.StatusOK {
			return true, nil
		}
	}

	return false, nil
}
func (ts *testService) WaitForCounterPodToUpdateValue() (bool, error) {
	count, err := ts.checkValue()
	if err != nil {
		log.Error(err)
		return false, err
	}

	if count != 1 {
		return false, nil
	}
	return true, nil
}

func (ts *testService) DeleteTestService() error {
	log.WithFields(log.Fields{"name": testServiceName}).Debug("Deleting Deployment")
	err := ts.K8sResourcesClient.DeleteDeployment(testServiceName, &v1.DeleteOptions{})

	if err != nil {
		log.Error(err)
		return err
	}

	log.WithFields(log.Fields{"name": testServiceName}).Debug("Deleting Service")
	err = ts.K8sResourcesClient.DeleteService(testServiceName, &v1.DeleteOptions{})

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (ts *testService) GetTestServiceURL() string {
	namespace := ts.K8sResourcesClient.GetNamespace()
	return fmt.Sprintf(counterEndpointFormat, testServiceName, namespace, testServicePort)
}

func (ts *testService) getHealthEndpointURL() string {
	namespace := ts.K8sResourcesClient.GetNamespace()
	return fmt.Sprintf(healthEndpointFormat, testServiceName, namespace, testServicePort)
}

func (ts *testService) createDeployment() error {
	log.WithFields(log.Fields{"name": testServiceName}).Debug("Creating Deployment")
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

func (ts *testService) createService() error {
	log.WithFields(log.Fields{"name": testServiceName}).Debug("Creating Service")
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
