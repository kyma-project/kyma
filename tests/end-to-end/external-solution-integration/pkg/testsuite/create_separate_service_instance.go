package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	acClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"time"
)

// CreateSeparateServiceInstance is a step which creates new separate ServiceInstances for API and Event
type CreateSeparateServiceInstance struct {
	serviceInstances serviceCatalogClient.ServiceInstanceInterface
	applications     acClient.ApplicationInterface
	state            CreateSeparateServiceInstanceState
	name             string
}

// CreateSeparateServiceInstanceState represents CreateServiceInstances dependencies
type CreateSeparateServiceInstanceState interface {
	SetAPIServiceInstanceName(string)
	SetEventServiceInstanceName(string)
	GetAPIServiceInstanceName() string
	GetEventServiceInstanceName() string
}

var _ step.Step = &CreateSeparateServiceInstance{}

// NewCreateSeparateServiceInstance returns new CreateSeparateServiceInstance
func NewCreateSeparateServiceInstance(name string, serviceInstances serviceCatalogClient.ServiceInstanceInterface, applications acClient.ApplicationInterface, state CreateSeparateServiceInstanceState) *CreateSeparateServiceInstance {
	return &CreateSeparateServiceInstance{
		name:             name,
		serviceInstances: serviceInstances,
		applications: applications,
		state:            state,
	}
}

// Name returns name of the step
func (s *CreateSeparateServiceInstance) Name() string {
	return "Create separate service instance for API and Event"
}

// Run executes the step
func (s *CreateSeparateServiceInstance) Run() error {
	scAPIExternalName, scEventExternalName, err := s.getServiceClassExternalNamesFromApplication()
	if err != nil {
		return err
	}

	siAPIName, err := s.createServiceInstance(scAPIExternalName, "api")
	if err != nil {
		return err
	}
	s.state.SetAPIServiceInstanceName(siAPIName)

	siEventName, err := s.createServiceInstance(scEventExternalName, "event")
	if err != nil {
		return err
	}
	s.state.SetEventServiceInstanceName(siEventName)

	return nil
}

func (s *CreateSeparateServiceInstance) getServiceClassExternalNamesFromApplication() (string, string, error) {
	var apiName, eventName string
	err := retry.Do(func() error {
		app, err := s.applications.Get(s.name, v1.GetOptions{})
		if err != nil {
			return err
		}
		for _, service := range app.Spec.Services {
			for _, entry := range service.Entries {
				switch entry.Type {
				case "Events":
					eventName = service.Name
				case "API":
					apiName = service.Name
				}
			}	
		}
		return nil
	})
	if apiName == "" || eventName == "" {
		return "", "", errors.New("service class for api or event not found")
	}
	return apiName, eventName, err
}

func (s *CreateSeparateServiceInstance) createServiceInstance(externalName, suffix string) (string, error) {
	si, err := s.serviceInstances.Create(&serviceCatalogApi.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:       fmt.Sprintf("%s-%s", s.name, suffix),
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: serviceCatalogApi.ServiceInstanceSpec{
			Parameters: &runtime.RawExtension{},
			PlanReference: serviceCatalogApi.PlanReference{
				ServiceClassExternalName: externalName,
				ServicePlanExternalName:  "default",
			},
			UpdateRequests: 0,
		},
	})
	if err != nil {
		return "", err
	}

	return si.Name, retry.Do(s.isServiceInstanceCreated(si.Name))
}

func (s *CreateSeparateServiceInstance) isServiceInstanceCreated(name string) func() error {
	return func() error {
		svcInstance, err := s.serviceInstances.Get(name, v1.GetOptions{})
		if err != nil {
			return err
		}

		if svcInstance.Status.ProvisionStatus != "Provisioned" {
			return errors.Errorf("unexpected provision status: %s", svcInstance.Status.ProvisionStatus)
		}
		return nil
	}
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateSeparateServiceInstance) Cleanup() error {
	err := s.serviceInstances.Delete(s.state.GetAPIServiceInstanceName(), &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = s.serviceInstances.Delete(s.state.GetEventServiceInstanceName(), &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		_, err := s.serviceInstances.Get(s.state.GetAPIServiceInstanceName(), v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		_, err = s.serviceInstances.Get(s.state.GetEventServiceInstanceName(), v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		return nil, nil
	}, retry.Delay(time.Second))
}
