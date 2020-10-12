package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InitializeConditions sets unset conditions to Unknown
func (s *SubscriptionStatus) InitializeConditions() {
	initialConditions := makeConditions()
	givenConditions := make(map[ConditionType]Condition, 0)

	// create map of Condition per ConditionType
	for _, condition := range s.Conditions {
		givenConditions[condition.Type] = condition
	}

	finalConditions := s.Conditions
	// check if every Condition is present in the current Conditions
	for _, expectedCondition := range initialConditions {
		if _, ok := givenConditions[expectedCondition.Type]; !ok {
			// and add it if it is missing
			finalConditions = append(finalConditions, expectedCondition)
		}
	}

	s.Conditions = finalConditions
}

// makeConditions creates an map of all conditions which the Subscription should have
func makeConditions() []Condition {
	conditions := []Condition{
		{
			Type:               ConditionSubscribed,
			LastTransitionTime: metav1.Now(),
			Status:             corev1.ConditionUnknown,
		},
	}
	return conditions
}
