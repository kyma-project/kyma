package testsuite

import (
	"github.com/pkg/errors"
	"github.com/avast/retry-go"
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type CreateServiceInstance struct {
	serviceInstances serviceCatalogClient.ServiceInstanceInterface
	state            CreateServiceInstanceState
}

type CreateServiceInstanceState interface {
	GetServiceID() string
}

var _ step.Step = &CreateServiceInstance{}

func NewCreateServiceInstance(serviceInstances serviceCatalogClient.ServiceInstanceInterface, state CreateServiceInstanceState) *CreateServiceInstance {
	return &CreateServiceInstance{
		serviceInstances: serviceInstances,
		state:            state,
	}
}

func (s *CreateServiceInstance) Name() string {
	return "Create service instance"
}

func (s *CreateServiceInstance) Run() error {
	siName := consts.ServiceInstanceName
	siID := consts.ServiceInstanceID
	_, err := s.serviceInstances.Create(&serviceCatalogApi.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:       siName,
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: serviceCatalogApi.ServiceInstanceSpec{
			ExternalID: siID,
			Parameters: &runtime.RawExtension{},
			PlanReference: serviceCatalogApi.PlanReference{
				ServiceClassName: s.state.GetServiceID(),
				ServicePlanName:  s.state.GetServiceID() + "-plan",
			},
			UpdateRequests: 0,
			UserInfo: &serviceCatalogApi.UserInfo{
				Groups: []string{
					"system:serviceaccounts",
					"system:serviceaccounts:kyma-system",
					"system:authenticated",
				},
				UID:      "",
				Username: "system:serviceaccount:kyma-system:core-console-backend-service",
			},
		},
	})
	if err != nil {
		return err
	}

	return retry.Do(s.isServiceInstanceCreated)
}

func (s *CreateServiceInstance) isServiceInstanceCreated() error {
	svcInstance, _ := s.serviceInstances.Get(consts.ServiceInstanceName, v1.GetOptions{})

	if svcInstance.Status.ProvisionStatus != "Provisioned" {
		return errors.New("Unexpected provision status: " + string(svcInstance.Status.ProvisionStatus))
	}
	return nil
}

func (s *CreateServiceInstance) Cleanup() error {
	siName := consts.ServiceInstanceName

	err := s.serviceInstances.Delete(siName, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return retry.Do(s.isServiceInstanceDeleted)
}

func (s *CreateServiceInstance) isServiceInstanceDeleted() error {
	_, err := s.serviceInstances.Get(consts.ServiceInstanceName, v1.GetOptions{})

	if err == nil {
		return errors.New("service instance still exists")
	}

	if !k8s_errors.IsNotFound(err) {
		return err
	}

	return nil
}
