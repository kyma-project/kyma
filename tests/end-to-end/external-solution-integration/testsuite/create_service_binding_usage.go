package testsuite

import (
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	serviceBindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	coreApi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreClient "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
)

type CreateServiceBindingUsage struct {
	*testkit.LambdaHelper
	serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface
	state                CreateServiceBindingUsageState
}

type CreateServiceBindingUsageState interface {
	GetServiceID() string
}

var _ step.Step = &CreateServiceBindingUsage{}

func NewCreateServiceBindingUsage(serviceBindingUsages serviceBindingUsageClient.ServiceBindingUsageInterface, pods coreClient.PodInterface, state CreateServiceBindingUsageState) *CreateServiceBindingUsage {
	return &CreateServiceBindingUsage{
		LambdaHelper:         testkit.NewLambdaHelper(pods),
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

	err = retry.Do(s.isLambdaBound, retry.Delay(300*time.Millisecond))
	if err != nil {
		return err
	}

	return nil
}

func (s *CreateServiceBindingUsage) isLambdaBound() error {
	sbuLabel := fmt.Sprintf("app-%s-%s", consts.AppName, s.state.GetServiceID())
	pods, err := s.ListLambdaPods()
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return errors.New("no function pods found")
	}

	for _, pod := range pods {
		if _, ok := pod.Labels[sbuLabel]; !ok {
			return errors.New("not bound pod exists: " + pod.Name)
		}

		for _, condition := range pod.Status.Conditions {
			if condition.Type == coreApi.PodReady && condition.Status != coreApi.ConditionTrue {
				return errors.New("pod not ready")
			}
		}
	}

	return nil
}

func (s *CreateServiceBindingUsage) Cleanup() error {
	return s.serviceBindingUsages.Delete(consts.ServiceBindingUsageName, &metav1.DeleteOptions{})
}
