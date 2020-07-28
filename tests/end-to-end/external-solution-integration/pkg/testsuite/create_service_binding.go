package testsuite

import (
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// CreateServiceBinding is a step which creates new ServiceBinding
type CreateServiceBinding struct {
	serviceBindings servicecatalogclientset.ServiceBindingInterface
	name            string
	serviceName     string
}

var _ step.Step = &CreateServiceBinding{}

// NewCreateServiceBinding returns new CreateServiceBinding
func NewCreateServiceBinding(name, serviceName string, serviceBindings servicecatalogclientset.ServiceBindingInterface) *CreateServiceBinding {
	return &CreateServiceBinding{
		serviceBindings: serviceBindings,
		name:            name,
		serviceName:     serviceName,
	}
}

// Name returns name name of the step
func (s *CreateServiceBinding) Name() string {
	return "Create service binding"
}

// Run executes the step
func (s *CreateServiceBinding) Run() error {
	serviceBinding := &scv1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{Name: s.name},
		Spec: scv1beta1.ServiceBindingSpec{
			InstanceRef: scv1beta1.LocalObjectReference{
				Name: s.serviceName,
			},
		},
	}

	_, err := s.serviceBindings.Create(serviceBinding)
	if err != nil {
		return err
	}

	return retry.Do(s.isServiceBindingReady)
}

func (s *CreateServiceBinding) isServiceBindingReady() error {
	sb, err := s.serviceBindings.Get(s.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, condition := range sb.Status.Conditions {
		if condition.Type == scv1beta1.ServiceBindingConditionReady {
			if condition.Status != scv1beta1.ConditionTrue {
				return errors.New("ServiceBinding is not ready")
			}
			break
		}
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateServiceBinding) Cleanup() error {
	err := s.serviceBindings.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		if e := s.forceCleanup(); e != nil {
			return errors.Wrapf(err, "while deleting service binding: %s with force cleanup error: %w", s.name, e)
		}
		return errors.Wrapf(err, "while deleting service binding: %s", s.name)
	}

	check := func() (interface{}, error) { return s.serviceBindings.Get(s.name, metav1.GetOptions{}) }
	if err := helpers.AwaitResourceDeleted(check); err == nil {
		return nil
	}

	return s.forceCleanup()
}

// forceCleanup clears the service binding finalizes list
// only if the service binding was marked for deletion and the finalizers list is not empty
func (s *CreateServiceBinding) forceCleanup() error {
	sb, err := s.serviceBindings.Get(s.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// skip cleanup if the service binding is not marked for deletion or the finalizers list is empty
	if sb.DeletionTimestamp == nil || len(sb.Finalizers) == 0 {
		return nil
	}

	sb = sb.DeepCopy()
	sb.Finalizers = []string{}
	if _, err = s.serviceBindings.Update(sb); err != nil {
		return err
	}

	return nil
}
