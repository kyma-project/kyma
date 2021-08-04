package process

import (
	"errors"
	"fmt"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
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

	s.process.State.FilteredSubscriptions = &eventingv1alpha1.SubscriptionList{
		TypeMeta: s.process.State.Subscriptions.TypeMeta,
		ListMeta: s.process.State.Subscriptions.ListMeta,
		Items:    []eventingv1alpha1.Subscription{},
	}

	subscriptionItems := s.process.State.Subscriptions.Items

	for _, subscription := range subscriptionItems {
		// generate the new name for the BEB webhook from subscription
		// @TODO: Replace MapSubscriptionName with actual method once available
		newSubscriptionName := MapSubscriptionName(&subscription)
		expectedConditionMessage := fmt.Sprintf("BEBId=%s", newSubscriptionName)
		conditionTypeToCheck := eventingv1alpha1.ConditionSubscriptionActive

		condition, err := s.findSubscriptionCondition(&subscription, conditionTypeToCheck)
		if err != nil {
			s.process.Logger.WithContext().Error(err)

			//@TODO: correct?
			// if condition not found, then we need to migrate this subscription
			s.process.State.FilteredSubscriptions.Items = append(s.process.State.FilteredSubscriptions.Items, subscription)

			continue
		}

		// Check the condition
		if string(condition.Message) == expectedConditionMessage {
			continue
		}

		// if reason dont match, then we need to migrate this subscription
		s.process.State.FilteredSubscriptions.Items = append(s.process.State.FilteredSubscriptions.Items, subscription)
	}



	//3) if not in condition, then check if

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

func MapSubscriptionName(sub *eventingv1alpha1.Subscription) string {
	// #TODO: Mocked function to be deleted later
	return sub.Name
}
