package testsuite

import (
	"fmt"
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateServiceBinding struct {
	serviceBindings serviceCatalogClient.ServiceBindingInterface
	endpoint        string
}

var _ step.Step = &CreateServiceBinding{}

func NewCreateServiceBinding(serviceBindings serviceCatalogClient.ServiceBindingInterface, namespace string) *CreateServiceBinding {
	return &CreateServiceBinding{
		serviceBindings: serviceBindings,
		endpoint:        fmt.Sprintf(consts.LambdaEndpointPattern, namespace),
	}
}

func (s *CreateServiceBinding) Name() string {
	return "Create service binding"
}

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

func (s *CreateServiceBinding) Cleanup() error {
	return s.serviceBindings.Delete(consts.ServiceBindingName, &metav1.DeleteOptions{})
}
