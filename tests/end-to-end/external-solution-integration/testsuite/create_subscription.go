package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	eventingApi "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eventingClient "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned/typed/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateSubscription struct {
	subscriptions eventingClient.SubscriptionInterface
	endpoint      string
}

var _ step.Step = &CreateSubscription{}

func NewCreateSubscription(subscriptions eventingClient.SubscriptionInterface, namespace string) *CreateSubscription {
	return &CreateSubscription{
		subscriptions: subscriptions,
		endpoint:      fmt.Sprintf(consts.LambdaEndpointPattern, namespace),
	}
}

func (s *CreateSubscription) Name() string {
	return "Create subscription"
}

func (s *CreateSubscription) Run() error {
	subSpec := eventingApi.SubscriptionSpec{
		Endpoint:                      s.endpoint,
		IncludeSubscriptionNameHeader: true,
		MaxInflight:                   400,
		PushRequestTimeoutMS:          2000,
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

func (s *CreateSubscription) Cleanup() error {
	return s.subscriptions.Delete(consts.AppName, &metav1.DeleteOptions{})
}

func (s *CreateSubscription) isSubscriptionReady() error {
	subscription, err := s.subscriptions.Get(consts.AppName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, condition := range subscription.Status.Conditions {
		if condition.Type == eventingApi.Ready && condition.Status != eventingApi.ConditionTrue {
			return errors.New("subscription not ready")
		}
	}
	return err
}
