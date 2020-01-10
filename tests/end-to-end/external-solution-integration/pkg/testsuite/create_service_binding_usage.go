package testsuite

import (
	"github.com/avast/retry-go"
	serviceBindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateLambdaServiceBindingUsage is a step which creates new ServiceBindingUsage
type CreateLambdaServiceBindingUsage struct {
	serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface
	name                 string
	serviceBindingName   string
	lambdaName           string
}

var _ step.Step = &CreateLambdaServiceBindingUsage{}

// NewCreateServiceBindingUsage returns new CreateLambdaServiceBindingUsage
func NewCreateServiceBindingUsage(name, serviceBindingName, lambdaName string, serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface) *CreateLambdaServiceBindingUsage {
	return &CreateLambdaServiceBindingUsage{
		serviceBindingUsages: serviceBindingUsages,
		name:                 name,
		serviceBindingName:   serviceBindingName,
		lambdaName:           lambdaName,
	}
}

// Name returns name name of the step
func (s *CreateLambdaServiceBindingUsage) Name() string {
	return "Create service binding usage"
}

// Run executes the step
func (s *CreateLambdaServiceBindingUsage) Run() error {
	serviceBindingUsage := &serviceBindingUsageApi.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{Kind: "ServiceBindingUsage", APIVersion: serviceBindingUsageApi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.name,
			Labels: map[string]string{"Function": s.lambdaName, "ServiceBinding": s.serviceBindingName},
		},
		Spec: serviceBindingUsageApi.ServiceBindingUsageSpec{
			Parameters: &serviceBindingUsageApi.Parameters{
				EnvPrefix: &serviceBindingUsageApi.EnvPrefix{
					Name: "",
				},
			},
			ServiceBindingRef: serviceBindingUsageApi.LocalReferenceByName{
				Name: s.serviceBindingName,
			},
			UsedBy: serviceBindingUsageApi.LocalReferenceByKindAndName{
				Kind: "function",
				Name: s.lambdaName,
			},
		},
	}

	_, err := s.serviceBindingUsages.Create(serviceBindingUsage)
	if err != nil {
		return err
	}

	return retry.Do(s.isServiceBindingUsageReady)
}

func (s *CreateLambdaServiceBindingUsage) isServiceBindingUsageReady() error {
	sbu, err := s.serviceBindingUsages.Get(s.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, condition := range sbu.Status.Conditions {
		if condition.Type == serviceBindingUsageApi.ServiceBindingUsageReady {
			if condition.Status != serviceBindingUsageApi.ConditionTrue {
				return errors.New("ServiceBinding is not ready")
			}
			break
		}
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateLambdaServiceBindingUsage) Cleanup() error {
	err := s.serviceBindingUsages.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.serviceBindingUsages.Get(s.serviceBindingName, metav1.GetOptions{})
	})
}
