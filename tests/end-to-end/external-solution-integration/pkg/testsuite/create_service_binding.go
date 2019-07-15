package testsuite

import (
	"fmt"
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateServiceBinding is a step which creates new ServiceBinding
type CreateServiceBinding struct {
	serviceBindings serviceCatalogClient.ServiceBindingInterface
	endpoint        string
}

var _ step.Step = &CreateServiceBinding{}

// NewCreateServiceBinding returns new CreateServiceBinding
func NewCreateServiceBinding(serviceBindings serviceCatalogClient.ServiceBindingInterface, namespace string) *CreateServiceBinding {
	return &CreateServiceBinding{
		serviceBindings: serviceBindings,
		endpoint:        fmt.Sprintf(consts.LambdaEndpointPattern, namespace),
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
			ExternalID: consts.ServiceBindingID,
			InstanceRef: serviceCatalogApi.LocalObjectReference{
				Name: consts.ServiceInstanceName,
			},
			SecretName: consts.ServiceBindingSecret,
			UserInfo: &serviceCatalogApi.UserInfo{
				Groups: []string{
					"system:authenticated",
				},
				UID:      "",
				Username: "adimn@kyma.cx",
			},
		},
	}

	_, err := s.serviceBindings.Create(serviceBinding)
	return err
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateServiceBinding) Cleanup() error {
	return s.serviceBindings.Delete(consts.ServiceBindingName, &metav1.DeleteOptions{})
}
