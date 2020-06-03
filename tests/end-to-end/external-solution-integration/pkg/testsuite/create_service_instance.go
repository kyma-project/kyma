package testsuite

import (
	"fmt"
	"time"

	retrygo "github.com/avast/retry-go"

	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// CreateServiceInstance is a step which creates new ServiceInstance
type CreateServiceInstance struct {
	serviceInstances servicecatalogclientset.ServiceInstanceInterface
	serviceClasses   servicecatalogclientset.ServiceClassInterface
	servicePlans     servicecatalogclientset.ServicePlanInterface
	name             string
	instanceName     string
	getClassIDFn     func() string
	getPlanIDFn      func() string
}

var _ step.Step = &CreateServiceInstance{}

// NewCreateServiceInstance returns new NewCreateServiceInstance
func NewCreateServiceInstance(name, instanceName string, getClassIDFn func() string, getPlanIDFn func() string, serviceInstances servicecatalogclientset.ServiceInstanceInterface, serviceClasses servicecatalogclientset.ServiceClassInterface, servicePlans servicecatalogclientset.ServicePlanInterface) *CreateServiceInstance {
	return &CreateServiceInstance{
		name:             name,
		instanceName:     instanceName,
		getClassIDFn:     getClassIDFn,
		getPlanIDFn:      getPlanIDFn,
		serviceInstances: serviceInstances,
		serviceClasses:   serviceClasses,
		servicePlans:     servicePlans,
	}
}

// Name returns name name of the step
func (s *CreateServiceInstance) Name() string {
	return fmt.Sprintf("Create service instance: %s", s.instanceName)
}

// Run executes the step
func (s *CreateServiceInstance) Run() error {
	serviceClassExternalName, err := s.findServiceClassExternalName(s.getClassIDFn())
	if err != nil {
		return err
	}

	servicePlanExternalName, err := s.findServicePlanExternalName(s.getPlanIDFn())
	if err != nil {
		return err
	}

	_, err = s.serviceInstances.Create(&scv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       s.instanceName,
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			Parameters: &runtime.RawExtension{},
			PlanReference: scv1beta1.PlanReference{
				ServiceClassExternalName: serviceClassExternalName,
				ServicePlanExternalName:  servicePlanExternalName,
			},
			UpdateRequests: 0,
		},
	})
	if err != nil {
		return err
	}

	return retry.Do(s.isServiceInstanceCreated, retrygo.Attempts(180), retrygo.Delay(time.Second))
}

func (s *CreateServiceInstance) findServiceClassExternalName(serviceClassID string) (string, error) {
	var name string
	err := retry.Do(func() error {
		serviceClass, err := s.serviceClasses.Get(serviceClassID, metav1.GetOptions{})
		if err != nil {
			return err
		}
		name = serviceClass.Spec.ExternalName
		return nil
	})
	return name, err
}

func (s *CreateServiceInstance) findServicePlanExternalName(servicePlanID string) (string, error) {
	var name string
	err := retry.Do(func() error {
		servicePlan, err := s.servicePlans.Get(servicePlanID, metav1.GetOptions{})
		if err != nil {
			return err
		}
		name = servicePlan.Spec.ExternalName
		return nil
	})
	return name, err
}

func (s *CreateServiceInstance) isServiceInstanceCreated() error {
	svcInstance, err := s.serviceInstances.Get(s.instanceName, metav1.GetOptions{})
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
	err := retry.Do(func() error {
		return s.serviceInstances.Delete(s.instanceName, &metav1.DeleteOptions{})
	})
	if err != nil {
		return errors.Wrapf(err, "while deleting service instance: %s", s.instanceName)
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.serviceInstances.Get(s.name, metav1.GetOptions{})
	})
}
