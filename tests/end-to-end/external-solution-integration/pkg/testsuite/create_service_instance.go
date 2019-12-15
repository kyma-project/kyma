package testsuite

import (
	"github.com/avast/retry-go"
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"time"
)

// CreateServiceInstance is a step which creates new ServiceInstance
type CreateServiceInstance struct {
	serviceInstances serviceCatalogClient.ServiceInstanceInterface
	serviceClasses   serviceCatalogClient.ServiceClassInterface
	state            CreateServiceInstanceState
	name             string
}

// CreateServiceInstanceState represents CreateServiceInstance dependencies
type CreateServiceInstanceState interface {
	GetServiceClassID() string
	SetAPIServiceInstanceName(string)
	SetEventServiceInstanceName(string)
	GetAPIServiceInstanceName() string
	GetEventServiceInstanceName() string
}

var _ step.Step = &CreateServiceInstance{}

// NewCreateServiceInstance returns new CreateServiceInstance
func NewCreateServiceInstance(name string, serviceInstances serviceCatalogClient.ServiceInstanceInterface, serviceClasses serviceCatalogClient.ServiceClassInterface, state CreateServiceInstanceState) *CreateServiceInstance {
	return &CreateServiceInstance{
		name:             name,
		serviceInstances: serviceInstances,
		serviceClasses:   serviceClasses,
		state:            state,
	}
}

// Name returns name name of the step
func (s *CreateServiceInstance) Name() string {
	return "Create service instance"
}

// Run executes the step
func (s *CreateServiceInstance) Run() error {
	scExternalName, err := s.findServiceClassExternalName()
	if err != nil {
		return err
	}

	si, err := s.serviceInstances.Create(&serviceCatalogApi.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:       s.name,
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: serviceCatalogApi.ServiceInstanceSpec{
			Parameters: &runtime.RawExtension{},
			PlanReference: serviceCatalogApi.PlanReference{
				ServiceClassExternalName: scExternalName,
				ServicePlanExternalName:  "default",
			},
			UpdateRequests: 0,
		},
	})
	if err != nil {
		return err
	}
	s.state.SetAPIServiceInstanceName(si.Name)
	s.state.SetEventServiceInstanceName(si.Name)

	return retry.Do(s.isServiceInstanceCreated)
}

func (s *CreateServiceInstance) findServiceClassExternalName() (string, error) {
	var name string
	err := retry.Do(func() error {
		sc, err := s.serviceClasses.Get(s.state.GetServiceClassID(), v1.GetOptions{})
		if err != nil {
			return err
		}
		name = sc.Spec.ExternalName
		return nil
	})
	return name, err
}

func (s *CreateServiceInstance) isServiceInstanceCreated() error {
	svcInstance, err := s.serviceInstances.Get(s.state.GetAPIServiceInstanceName(), v1.GetOptions{})
	if err != nil {
		return err
	}

	if svcInstance.Status.ProvisionStatus != "Provisioned" {
		return errors.Errorf("unexpected provision status: %s", svcInstance.Status.ProvisionStatus)
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateServiceInstance) Cleanup() error {
	err := s.serviceInstances.Delete(s.name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}
	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.serviceInstances.Get(s.name, v1.GetOptions{})
	}, retry.Delay(time.Second))
}
