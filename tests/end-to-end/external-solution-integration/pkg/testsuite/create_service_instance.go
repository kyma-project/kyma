package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	serviceCatalogApi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"time"
)

type CreateServiceInstance struct {
	serviceInstances serviceCatalogClient.ServiceInstanceInterface
	state            CreateServiceInstanceState
	runID            string
}

type CreateServiceInstanceState interface {
	GetServiceClassID() string
	SetServiceInstanceName(string)
	GetServiceInstanceName() string
}

var _ step.Step = &CreateServiceInstance{}

func NewCreateServiceInstance(runID string, serviceInstances serviceCatalogClient.ServiceInstanceInterface, state CreateServiceInstanceState) *CreateServiceInstance {
	return &CreateServiceInstance{
		runID:            runID,
		serviceInstances: serviceInstances,
		state:            state,
	}
}

func (s *CreateServiceInstance) Name() string {
	return "Create service instance"
}

func (s *CreateServiceInstance) Run() error {
	si, err := s.serviceInstances.Create(&serviceCatalogApi.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: consts.ServiceInstanceName,
			Finalizers:   []string{"kubernetes-incubator/service-catalog"},
			Labels:       map[string]string{"runID": s.runID},
		},
		Spec: serviceCatalogApi.ServiceInstanceSpec{
			Parameters: &runtime.RawExtension{},
			PlanReference: serviceCatalogApi.PlanReference{
				ServiceClassName: s.state.GetServiceClassID(),
				ServicePlanName:  s.state.GetServiceClassID() + "-plan",
			},
			UpdateRequests: 0,
		},
	})
	if err != nil {
		return err
	}
	s.state.SetServiceInstanceName(si.Name)

	return retry.Do(s.isServiceInstanceCreated)
}

func (s *CreateServiceInstance) isServiceInstanceCreated() error {
	svcInstance, err := s.serviceInstances.Get(s.state.GetServiceInstanceName(), v1.GetOptions{})
	if err != nil {
		return err
	}

	if svcInstance.Status.ProvisionStatus != "Provisioned" {
		return errors.Errorf("unexpected provision status: %s", svcInstance.Status.ProvisionStatus)
	}
	return nil
}

func (s *CreateServiceInstance) Cleanup() error {
	selector := v1.ListOptions{
		LabelSelector: fmt.Sprintf("runID=%s", s.runID),
	}
	err := s.serviceInstances.DeleteCollection(&v1.DeleteOptions{}, selector)
	if err != nil {
		return err
	}
	return retry.Do(func() error {
		list, err := s.serviceInstances.List(selector)
		if err != nil {
			return err
		}

		if len(list.Items) != 0 {
			return errors.New("service instance still exists")
		}

		return nil
	}, retry.Delay(time.Second))
}
