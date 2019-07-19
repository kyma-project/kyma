package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	serviceBindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreClient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

// CreateLambdaServiceBindingUsage is a step which creates new ServiceBindingUsage
type CreateLambdaServiceBindingUsage struct {
	*helpers.LambdaHelper
	serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface
	state                CreateServiceBindingUsageState
	name                 string
	serviceBindingName   string
	lambdaName           string
}

// CreateServiceBindingUsageState represents CreateLambdaServiceBindingUsage dependencies
type CreateServiceBindingUsageState interface {
	GetServiceClassID() string
}

var _ step.Step = &CreateLambdaServiceBindingUsage{}

// NewCreateServiceBindingUsage returns new CreateLambdaServiceBindingUsage
func NewCreateServiceBindingUsage(name, serviceBindingName, lambdaName string, serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface, pods coreClient.PodInterface, state CreateServiceBindingUsageState) *CreateLambdaServiceBindingUsage {
	return &CreateLambdaServiceBindingUsage{
		LambdaHelper:         helpers.NewLambdaHelper(pods),
		serviceBindingUsages: serviceBindingUsages,
		state:                state,
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

	err = retry.Do(s.isLambdaBound, retry.Delay(500*time.Millisecond))
	if err != nil {
		return err
	}

	return nil
}

func (s *CreateLambdaServiceBindingUsage) isLambdaBound() error {
	sbuLabel := fmt.Sprintf("app-%s-%s", s.lambdaName, s.state.GetServiceClassID())
	pods, err := s.ListLambdaPods(s.lambdaName)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return errors.New("no function pods found")
	}

	for _, pod := range pods {
		if _, ok := pod.Labels[sbuLabel]; !ok {
			return errors.Errorf("not bound pod exists: %s", pod.Name)
		}

		if !helpers.IsPodReady(pod) {
			return errors.New("pod is not ready yet")
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
