package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	serviceBindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreClient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

// CreateServiceBindingUsage is a step which creates new ServiceBindingUsage
type CreateServiceBindingUsage struct {
	*helpers.LambdaHelper
	serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface
	state                CreateServiceBindingUsageState
}

type CreateServiceBindingUsageState interface {
	GetServiceClassID() string
}

var _ step.Step = &CreateServiceBindingUsage{}

func NewCreateServiceBindingUsage(serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface, pods coreClient.PodInterface, state CreateServiceBindingUsageState) *CreateServiceBindingUsage {
	return &CreateServiceBindingUsage{
		LambdaHelper:         helpers.NewLambdaHelper(pods),
		serviceBindingUsages: serviceBindingUsages,
		state:                state,
	}
}

func (s *CreateServiceBindingUsage) Name() string {
	return "Create service binding usage"
}

func (s *CreateServiceBindingUsage) Run() error {
	serviceBindingUsage := &serviceBindingUsageApi.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{Kind: "ServiceBindingUsage", APIVersion: serviceBindingUsageApi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:   consts.ServiceBindingUsageName,
			Labels: map[string]string{"Function": consts.AppName, "ServiceBinding": consts.ServiceBindingName},
		},
		Spec: serviceBindingUsageApi.ServiceBindingUsageSpec{
			Parameters: &serviceBindingUsageApi.Parameters{
				EnvPrefix: &serviceBindingUsageApi.EnvPrefix{
					Name: "",
				},
			},
			ServiceBindingRef: serviceBindingUsageApi.LocalReferenceByName{
				Name: consts.ServiceBindingName,
			},
			UsedBy: serviceBindingUsageApi.LocalReferenceByKindAndName{
				Kind: "function",
				Name: consts.AppName,
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

func (s *CreateServiceBindingUsage) isLambdaBound() error {
	sbuLabel := fmt.Sprintf("app-%s-%s", consts.AppName, s.state.GetServiceClassID())
	pods, err := s.ListLambdaPods()
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

func (s *CreateServiceBindingUsage) Cleanup() error {
	err := s.serviceBindingUsages.Delete(consts.ServiceBindingUsageName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.serviceBindingUsages.Get(consts.ServiceBindingName, metav1.GetOptions{})
	})
}
