package process

import (
	"fmt"
	"strings"

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

	// First set the subscription name mapper
	s.process.Logger.WithContext().Info(fmt.Sprintf("Using domain name: %s in step: %s", s.process.Domain, s.ToString()))
	nameMapper := handlers.NewBebSubscriptionNameMapper(strings.TrimSpace(s.process.Domain), handlers.MaxBEBSubscriptionNameLength)

	// Now filter out the subscriptions which are not migrated
	s.process.State.FilteredSubscriptions = &eventingv1alpha1.SubscriptionList{
		TypeMeta: s.process.State.Subscriptions.TypeMeta,
		ListMeta: s.process.State.Subscriptions.ListMeta,
		Items:    []eventingv1alpha1.Subscription{},
	}

	for _, subscription := range s.process.State.Subscriptions.Items {
		// generate the new name for the BEB webhook from subscription
		newBebSubscriptionName := nameMapper.MapSubscriptionName(&subscription)
		expectedConditionMessage := eventingv1alpha1.CreateMessageForConditionReasonSubscriptionCreated(newBebSubscriptionName)
		conditionTypeToCheck := eventingv1alpha1.ConditionSubscribed

		condition, err := s.findSubscriptionCondition(&subscription, conditionTypeToCheck)

		// If condition found and has expected message
		// then we don't need to migrate
		if err == nil && condition.Message == expectedConditionMessage {
			continue
		}

		if err != nil {
			s.process.Logger.WithContext().Warn(err)
		}
		// Append to Filtered Subscriptions list, which needs to be migrated
		s.process.State.FilteredSubscriptions.Items = append(s.process.State.FilteredSubscriptions.Items, subscription)
	}

	s.process.Logger.WithContext().Info(fmt.Sprintf("Total Subscriptions: %d, Unmigrated Subscriptions: %d", len(s.process.State.Subscriptions.Items), len(s.process.State.FilteredSubscriptions.Items)))
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
	return nil, fmt.Errorf("failed to find condition with type: %s in subscription: %s", conditionType, subscription.Name)
}
