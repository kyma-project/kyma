package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	eventingApi "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eventingClient "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned/typed/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateSubscription is a step which creates new Subscription
type CreateSubscription struct {
	subscriptions eventingClient.SubscriptionInterface
	endpoint      string
}

var _ step.Step = &CreateSubscription{}

// NewCreateSubscription returns new CreateSubscription
func NewCreateSubscription(subscriptions eventingClient.SubscriptionInterface, namespace string) *CreateSubscription {
	return &CreateSubscription{
		subscriptions: subscriptions,
		endpoint:      fmt.Sprintf(consts.LambdaEndpointPattern, namespace),
	}
}

// Name returns name name of the step
func (s *CreateSubscription) Name() string {
	return "Create subscription"
}

// Run executes the step
func (s *CreateSubscription) Run() error {
	subSpec := eventingApi.SubscriptionSpec{
		Endpoint:                      s.endpoint,
		IncludeSubscriptionNameHeader: true,
		EventType:                     consts.EventType,
		EventTypeVersion:              consts.EventVersion,
		SourceID:                      consts.AppName,
	}

	sub := &eventingApi.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:   consts.AppName,
			Labels: map[string]string{"Function": consts.AppName},
		},

		SubscriptionSpec: subSpec,
	}

	_, err := s.subscriptions.Create(sub)
	if err != nil {
		return err
	}

	return retry.Do(s.isSubscriptionReady)
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateSubscription) Cleanup() error {
	return s.subscriptions.Delete(consts.AppName, &metav1.DeleteOptions{})
}

func (s *CreateSubscription) isSubscriptionReady() error {
	subscription, err := s.subscriptions.Get(consts.AppName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, condition := range subscription.Status.Conditions {
		if condition.Status != eventingApi.ConditionTrue {
			return errors.Errorf("subscription condition not true: %s", condition.Type)
		}
	}
	return nil
}
