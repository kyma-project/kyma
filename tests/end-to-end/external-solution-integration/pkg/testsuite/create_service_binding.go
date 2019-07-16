package testsuite

import (
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateServiceBinding is a step which creates new ServiceBinding
type CreateServiceBinding struct {
	testkit.K8sHelper
	serviceBindings serviceCatalogClient.ServiceBindingInterface
	state           CreateServiceBindingState
}

type CreateServiceBindingState interface {
	GetServiceInstanceName() string
}

var _ step.Step = &CreateServiceBinding{}

// NewCreateServiceBinding returns new CreateServiceBinding
func NewCreateServiceBinding(serviceBindings serviceCatalogClient.ServiceBindingInterface, state CreateServiceBindingState) *CreateServiceBinding {
	return &CreateServiceBinding{
		serviceBindings: serviceBindings,
		state:           state,
	}
}

// Name returns name name of the step
func (s *CreateServiceBinding) Name() string {
	return "Create service binding"
}

// Run executes the step
func (s *CreateServiceBinding) Run() error {
	serviceBinding := &serviceCatalogApi.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{Name: consts.ServiceBindingName},
		Spec: serviceCatalogApi.ServiceBindingSpec{
			InstanceRef: serviceCatalogApi.LocalObjectReference{
				Name: s.state.GetServiceInstanceName(),
			},
			SecretName: consts.ServiceBindingSecret,
		},
	}

	_, err := s.serviceBindings.Create(serviceBinding)
	return err
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateServiceBinding) Cleanup() error {
	err := s.serviceBindings.Delete(consts.ServiceBindingName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return s.AwaitResourceDeleted(func() (interface{}, error) {
		return s.serviceBindings.Get(consts.ServiceBindingName, metav1.GetOptions{})
	})
}
