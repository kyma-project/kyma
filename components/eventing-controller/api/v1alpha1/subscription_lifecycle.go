package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
)

type ConditionType string

const (
	ConditionSubscribed ConditionType = "Subscribed"
)

type Condition struct {
	Type               ConditionType      `json:"type,omitempty"`
	Status             v1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason    `json:"reason,omitempty"`
	Message            string             `json:"message,omitempty"`
}

type ConditionReason string

const (
	ConditionReasonSubscriptionCreated ConditionReason = "SubscriptionCreated"
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

func MakeCondition(conditionType ConditionType, reason ConditionReason, status corev1.ConditionStatus) Condition {
	return Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		// TODO:
		Message: "",
	}
}

func (s *SubscriptionStatus) IsConditionSubscribed() bool {
	for _, condition := range s.Conditions {
		if condition.Type == ConditionSubscribed && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
