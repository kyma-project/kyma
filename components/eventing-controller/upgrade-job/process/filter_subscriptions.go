package process

import (
	"errors"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
)

var _ Step = &FilterSubscriptions{}

// FilterSubscriptions struct implements the interface Step
type FilterSubscriptions struct {
	name    string
	process *Process
}

// NewFilterSubscriptions returns new instance of FilterSubscriptions struct
func NewFilterSubscriptions(p *Process) FilterSubscriptions {
	return FilterSubscriptions{
		name:    "Filter subscriptions based on migration",
		process: p,
	}
}

// ToString returns step name
func (s FilterSubscriptions) ToString() string {
	return s.name
}

// Do filters subscriptions based on if they were already migrated or not
// by checking if the new BEB webhook name is in subscription condition
func (s FilterSubscriptions) Do() error {

	// First get the shootName, and initialize BebSubscriptionNameMapper
	shootName := ""
	configmap, err := s.process.Clients.ConfigMap.Get("kube-system", "shoot-info")
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	if err == nil {
		shootName = configmap.Data["shootName"]
	}

	nameMapper := handlers.NewBebSubscriptionNameMapper(shootName, handlers.MaxBEBSubscriptionNameLength)

	// Now filter out the subscriptions which are not migrated
	s.process.State.FilteredSubscriptions = &eventingv1alpha1.SubscriptionList{
		TypeMeta: s.process.State.Subscriptions.TypeMeta,
		ListMeta: s.process.State.Subscriptions.ListMeta,
		Items:    []eventingv1alpha1.Subscription{},
	}

	var subscriptionItems []eventingv1alpha1.Subscription
	if s.process.State.Subscriptions != nil {
		subscriptionItems = s.process.State.Subscriptions.Items
	}

	for _, subscription := range subscriptionItems {
		// generate the new name for the BEB webhook from subscription
		newBebSubscriptionName := nameMapper.MapSubscriptionName(&subscription)
		expectedConditionMessage := eventingv1alpha1.CreateMessageForConditionReasonSubscriptionCreated(newBebSubscriptionName)
		conditionTypeToCheck := eventingv1alpha1.ConditionSubscribed

		condition, err := s.findSubscriptionCondition(&subscription, conditionTypeToCheck)
		if err != nil {
			s.process.Logger.WithContext().Error(err)

			// if condition not found, then we need to migrate this subscription
			s.process.State.FilteredSubscriptions.Items = append(s.process.State.FilteredSubscriptions.Items, subscription)
			continue
		}

		// If the condition message don't match with expectedConditionMessage
		// then we need to migrate this subscription
		if string(condition.Message) != expectedConditionMessage {
			s.process.State.FilteredSubscriptions.Items = append(s.process.State.FilteredSubscriptions.Items, subscription)
		}
	}

	return nil
}

// findSubscriptionCondition returns condition object matching the provided conditionType from subscription, if found
// or returns error if condition not found
func (s FilterSubscriptions) findSubscriptionCondition(subscription *eventingv1alpha1.Subscription, conditionType eventingv1alpha1.ConditionType) (*eventingv1alpha1.Condition, error) {
	for _, condition := range subscription.Status.Conditions {
		if condition.Type == conditionType {
			return &condition, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("failed to find condition with type: %s in subscription: %s", conditionType, subscription.Name))
}
