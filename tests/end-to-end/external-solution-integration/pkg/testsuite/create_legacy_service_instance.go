package testsuite

import (
	"fmt"

	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// CreateLegacyServiceInstance is a step which creates new ServiceInstance
type CreateLegacyServiceInstance struct {
	serviceInstances servicecatalogclientset.ServiceInstanceInterface
	serviceClasses   servicecatalogclientset.ServiceClassInterface
	name             string
	instanceName     string
	getClassIDFn     func() string
}

var _ step.Step = &CreateLegacyServiceInstance{}

// NewCreateLegacyServiceInstance returns new CreateLegacyServiceInstance
func NewCreateLegacyServiceInstance(name, instanceName string, get func() string, serviceInstances servicecatalogclientset.ServiceInstanceInterface, serviceClasses servicecatalogclientset.ServiceClassInterface) *CreateLegacyServiceInstance {
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

	_, err = s.serviceInstances.Create(&scv1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:       s.instanceName,
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			Parameters: &runtime.RawExtension{},
			PlanReference: scv1beta1.PlanReference{
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
	err := retry.Do(func() error { return s.serviceInstances.Delete(s.name, &v1.DeleteOptions{}) })
	if err != nil {
		if e := s.forceCleanup(); e != nil {
			return errors.Wrapf(err, "while deleting service instance: %s with force cleanup error: %w", s.name, e)
		}
		return errors.Wrapf(err, "while deleting service instance: %s", s.name)
	}

	check := func() (interface{}, error) { return s.serviceInstances.Get(s.name, v1.GetOptions{}) }
	if err := helpers.AwaitResourceDeleted(check); err == nil {
		return nil
	}

	return s.forceCleanup()
}

// forceCleanup clears the service instance finalizes list
// only if the service instance was marked for deletion and the finalizers list is not empty
func (s *CreateLegacyServiceInstance) forceCleanup() error {
	si, err := s.serviceInstances.Get(s.name, v1.GetOptions{})
	if err != nil {
		return err
	}

	// skip cleanup if the service instance is not marked for deletion or the finalizers list is empty
	if si.DeletionTimestamp == nil || len(si.Finalizers) == 0 {
		return nil
	}

	si = si.DeepCopy()
	si.Finalizers = []string{}
	if _, err = s.serviceInstances.Update(si); err != nil {
		return err
	}

	return nil
}
