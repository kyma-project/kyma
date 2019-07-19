package testsuite

import (
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateServiceBinding is a step which creates new ServiceBinding
type CreateServiceBinding struct {
	serviceBindings serviceCatalogClient.ServiceBindingInterface
	state           CreateServiceBindingState
	name            string
	secretName      string
}

// CreateServiceBindingState represents CreateServiceBinding dependencies
type CreateServiceBindingState interface {
	GetServiceInstanceName() string
}

var _ step.Step = &CreateServiceBinding{}

// NewCreateServiceBinding returns new CreateServiceBinding
func NewCreateServiceBinding(name, secretName string, serviceBindings serviceCatalogClient.ServiceBindingInterface, state CreateServiceBindingState) *CreateServiceBinding {
	return &CreateServiceBinding{
		serviceBindings: serviceBindings,
		state:           state,
		name:            name,
		secretName:      secretName,
	}
}

// Name returns name name of the step
func (s *CreateServiceBinding) Name() string {
	return "Create service binding"
}

// Run executes the step
func (s *CreateServiceBinding) Run() error {
	serviceBinding := &serviceCatalogApi.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{Name: s.name},
		Spec: serviceCatalogApi.ServiceBindingSpec{
			InstanceRef: serviceCatalogApi.LocalObjectReference{
				Name: s.state.GetServiceInstanceName(),
			},
			SecretName: s.secretName,
		},
	}

	_, err := s.serviceBindings.Create(serviceBinding)
	return err
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateServiceBinding) Cleanup() error {
	err := s.serviceBindings.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.serviceBindings.Get(s.name, metav1.GetOptions{})
	})
}
