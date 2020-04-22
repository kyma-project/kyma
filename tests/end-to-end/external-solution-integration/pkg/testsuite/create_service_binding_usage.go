package testsuite

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	eventingv1alpha1clientset "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingv1alpha1clientset "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"

	sbuv1alpha1 "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	sbuclientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

const (
	applicationName = "application-name"
	brokerNamespace = "broker-namespace"
)

// CreateLambdaServiceBindingUsage is a step which creates new ServiceBindingUsage
type CreateLambdaServiceBindingUsage struct {
	serviceBindingUsages sbuclientset.ServiceBindingUsageInterface
	broker               eventingv1alpha1clientset.BrokerInterface
	subscription         messagingv1alpha1clientset.SubscriptionInterface
	name                 string
	serviceBindingName   string
	lambdaName           string
}

var _ step.Step = &CreateLambdaServiceBindingUsage{}

// NewCreateServiceBindingUsage returns new CreateLambdaServiceBindingUsage
func NewCreateServiceBindingUsage(name, serviceBindingName, lambdaName string,
	serviceBindingUsages sbuclientset.ServiceBindingUsageInterface,
	knativeBroker eventingv1alpha1clientset.BrokerInterface,
	knativeSubscription messagingv1alpha1clientset.SubscriptionInterface) *CreateLambdaServiceBindingUsage {
	return &CreateLambdaServiceBindingUsage{
		serviceBindingUsages: serviceBindingUsages,
		broker:               knativeBroker,
		subscription:         knativeSubscription,
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
	serviceBindingUsage := &sbuv1alpha1.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{Kind: "ServiceBindingUsage", APIVersion: sbuv1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.name,
			Labels: map[string]string{"Function": s.lambdaName, "ServiceBinding": s.serviceBindingName},
		},
		Spec: sbuv1alpha1.ServiceBindingUsageSpec{
			Parameters: &sbuv1alpha1.Parameters{
				EnvPrefix: &sbuv1alpha1.EnvPrefix{
					Name: "",
				},
			},
			ServiceBindingRef: sbuv1alpha1.LocalReferenceByName{
				Name: s.serviceBindingName,
			},
			UsedBy: sbuv1alpha1.LocalReferenceByKindAndName{
				Kind: "deployment",
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
		if condition.Type == sbuv1alpha1.ServiceBindingUsageReady {
			if condition.Status != sbuv1alpha1.ConditionTrue {
				return errors.New("ServiceBinding is not ready")
			}
			break
		}
	}

	knativeSubscriptionLabels := make(map[string]string)
	knativeSubscriptionLabels[applicationName] = s.name
	knativeSubscriptionLabels[brokerNamespace] = s.name

	if s.subscription != nil {
		knativeSubscription, err := s.subscription.List(metav1.ListOptions{
			LabelSelector: labels.Set(knativeSubscriptionLabels).String(),
		})
		if err != nil {
			return err
		}

		if length := len(knativeSubscription.Items); length == 0 || length > 1 {
			return errors.Errorf("unexpected number of knative subscriptions were found.\n Number of knative Subscriptions: %d", length)
		}
	}

	return retry.Do(s.isBrokerReady)
}

func (s *CreateLambdaServiceBindingUsage) isBrokerReady() error {
	if s.broker != nil {
		knativeBroker, err := s.broker.Get("default", metav1.GetOptions{})
		if err != nil {
			return err
		}

		if !knativeBroker.Status.IsReady() {
			return errors.Errorf("default knative broker in %s namespace is not ready. Status of Knative Broker: \n %+v", knativeBroker.Namespace, knativeBroker.Status)
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
