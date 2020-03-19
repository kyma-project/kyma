package testsuite

import (
	"fmt"

	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CreateLegacyServiceInstance is a step which creates new ServiceInstance
type CreateLegacyServiceInstance struct {
	serviceInstances serviceCatalogClient.ServiceInstanceInterface
	serviceClasses   serviceCatalogClient.ServiceClassInterface
	name             string
	instanceName     string
	getClassIDFn     func() string
}

var _ step.Step = &CreateLegacyServiceInstance{}

// NewCreateLegacyServiceInstance returns new CreateLegacyServiceInstance
func NewCreateLegacyServiceInstance(name, instanceName string, get func() string, serviceInstances serviceCatalogClient.ServiceInstanceInterface, serviceClasses serviceCatalogClient.ServiceClassInterface) *CreateLegacyServiceInstance {
	return &CreateLegacyServiceInstance{
		name:             name,
		instanceName:     instanceName,
		getClassIDFn:     get,
		serviceInstances: serviceInstances,
		serviceClasses:   serviceClasses,
	}
}

// Name returns name name of the step
func (s *CreateLegacyServiceInstance) Name() string {
	return fmt.Sprintf("Create service instance: %s", s.instanceName)
}

// Run executes the step
func (s *CreateLegacyServiceInstance) Run() error {
	scExternalName, err := s.findServiceClassExternalName(s.getClassIDFn())
	if err != nil {
		return err
	}

	_, err = s.serviceInstances.Create(&serviceCatalogApi.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:       s.instanceName,
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
	return retry.Do(s.isServiceInstanceCreated)
}

func (s *CreateLegacyServiceInstance) findServiceClassExternalName(serviceClassID string) (string, error) {
	var name string
	err := retry.Do(func() error {
		sc, err := s.serviceClasses.Get(serviceClassID, v1.GetOptions{})
		if err != nil {
			return err
		}
		name = sc.Spec.ExternalName
		return nil
	})
	return name, err
}

func (s *CreateLegacyServiceInstance) isServiceInstanceCreated() error {
	svcInstance, err := s.serviceInstances.Get(s.instanceName, v1.GetOptions{})
	if err != nil {
		return err
	}

	if svcInstance.Status.ProvisionStatus != "Provisioned" {
		return errors.Errorf("unexpected provision status: %s", svcInstance.Status.ProvisionStatus)
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateLegacyServiceInstance) Cleanup() error {
	err := retry.Do(func() error {
		return s.serviceInstances.Delete(s.instanceName, &v1.DeleteOptions{})
	})
	if err != nil {
		return errors.Wrapf(err, "while deleting service instance: %s", s.instanceName)
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.serviceInstances.Get(s.name, v1.GetOptions{})
	})
}
