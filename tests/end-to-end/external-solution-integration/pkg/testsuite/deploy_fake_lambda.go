package testsuite

import (
	"fmt"
	"strconv"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

const (
	LegacyEnvKey          = "LEGACY"
	ExpectedPayloadEnvKey = "EXPECTED_PAYLOAD"
	image                 = "eu.gcr.io/kyma-project/fake-lambda:PR-7800"
)

// DeployFakeLambda deploys lambda to the cluster. The lambda will do PUT /counter to connected application upon receiving
// an event
type DeployFakeLambda struct {
	deployment      appsclient.DeploymentInterface
	service         coreclient.ServiceInterface
	pod             coreclient.PodInterface
	name            string
	port            int
	expectedPayload string
	legacy          string
}

var _ step.Step = &DeployFakeLambda{}

// NewDeployFakeLambda returns new DeployFakeLambda
func NewDeployFakeLambda(
	name, expectedPayload string, port int,
	deployment appsclient.DeploymentInterface, service coreclient.ServiceInterface,
	pod coreclient.PodInterface, legacy bool) *DeployFakeLambda {
	return &DeployFakeLambda{
		deployment:      deployment,
		service:         service,
		pod:             pod,
		name:            name,
		port:            port,
		expectedPayload: expectedPayload,
		legacy:          strconv.FormatBool(legacy),
	}
}

// Name returns name name of the step
func (s *DeployFakeLambda) Name() string {
	return fmt.Sprintf("Deploy fake lambda %s", s.name)
}

// Run executes the step
func (s *DeployFakeLambda) Run() error {
	deployment := s.fixDeployment()
	_, err := s.deployment.Create(deployment)
	if err != nil {
		return err
	}
	err = retry.Do(s.isDeploymentReady)
	if err != nil {
		return errors.Wrap(err, "deployment not ready")
	}

	service := s.fixService()
	_, err = s.service.Create(service)
	if err != nil {
		return err
	}

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *DeployFakeLambda) Cleanup() error {
	err := s.service.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = s.deployment.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	if err := retry.Do(s.isDeploymentTerminated); err != nil {
		return err
	}
	return retry.Do(s.isServiceTerminated)
}

func (s *DeployFakeLambda) fixService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.name,
			Labels: s.fixLabels(),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "http-function-port",
					Port:       8080,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(s.port),
				},
			},
			Selector: s.fixSelector(),
		},
	}
}

func (s *DeployFakeLambda) fixDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.name,
			Labels: s.fixLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: s.fixSelector(),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   s.name,
					Labels: s.fixLabels(),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  s.name,
							Image: image,
							Env: []v1.EnvVar{
								{
									Name:  LegacyEnvKey,
									Value: s.legacy,
								},
								{
									Name:  ExpectedPayloadEnvKey,
									Value: s.expectedPayload,
								},
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: int32(s.port),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (s *DeployFakeLambda) fixLabels() map[string]string {
	return map[string]string{"created-by": "kubeless", "function": s.name, "app": s.name}
}

func (s *DeployFakeLambda) fixSelector() map[string]string {
	return map[string]string{"app": s.name}
}

func (s *DeployFakeLambda) fixListOptions() metav1.ListOptions {
	return metav1.ListOptions{LabelSelector: labels.SelectorFromSet(s.fixSelector()).String()}
}

func (s *DeployFakeLambda) isDeploymentReady() error {
	deploymentList, err := s.deployment.List(s.fixListOptions())
	if err != nil {
		return err
	}

	if len(deploymentList.Items) == 0 {
		return errors.New("no deployment pods found")
	}

	for _, deployment := range deploymentList.Items {
		if !helpers.IsDeploymentReady(deployment) {
			return errors.New("deployment is not ready yet")
		}
	}

	podList, err := s.pod.List(s.fixListOptions())
	if err != nil {
		return err
	}

	if len(podList.Items) != 0 {
		return errors.New("deployment pods not found")
	}

	for _, pod := range podList.Items {
		if !helpers.IsPodReady(pod) {
			return errors.New("pod is not ready yet")
		}
	}

	return nil
}

func (s *DeployFakeLambda) isDeploymentTerminated() error {
	deploymentList, err := s.deployment.List(s.fixListOptions())
	if err != nil {
		return err
	}

	if len(deploymentList.Items) != 0 {
		return errors.New("deployment not found")
	}

	return nil
}

func (s *DeployFakeLambda) isServiceTerminated() error {
	serviceList, err := s.service.List(s.fixListOptions())
	if err != nil {
		return err
	}

	if len(serviceList.Items) != 0 {
		return errors.New("function pods found")
	}

	return nil
}
