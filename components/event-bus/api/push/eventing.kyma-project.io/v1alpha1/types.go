package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//Subscription describes a subscription
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	SubscriptionSpec  `json:"spec"`
	Status            SubscriptionStatus `json:"status,omitempty"`
}

// SubscriptionSpec for Event Bus Push
type SubscriptionSpec struct {
	Endpoint                      string `json:"endpoint"`
	IncludeSubscriptionNameHeader bool   `json:"include_subscription_name_header"`
	EventType                     string `json:"event_type"`
	EventTypeVersion              string `json:"event_type_version"`
	SourceID                      string `json:"source_id"`
}

// SubscriptionList of Kyma subscriptions.
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Subscription `json:"items"`
}

// SubscriptionStatus of Kyma subscriptions.
type SubscriptionStatus struct {
	Status `json:",inline"`
}

// Status of Kyma subscriptions.
type Status struct {
	Conditions Conditions `json:"conditions"`
}

// SubscriptionCondition of Kyma subscriptions.
type SubscriptionCondition struct {
	Type               SubscriptionConditionType `json:"type"`
	Status             ConditionStatus           `json:"status"`
	LastTransitionTime metav1.Time               `json:"last_transition_time"`
	Reason             string                    `json:"reason"`
	Message            string                    `json:"message"`
}

// ConditionStatus type
type ConditionStatus string

// SubscriptionConditionType type
type SubscriptionConditionType string

// Conditions type
type Conditions []SubscriptionCondition

const (
	// ConditionTrue value for the Kyma subscription condition status.
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse value for the Kyma subscription condition status.
	ConditionFalse ConditionStatus = "False"

	// Ready label for the Kyma subscription condition type.
	Ready SubscriptionConditionType = "is-ready"
	// EventsActivated label for the Kyma subscription condition type.
	EventsActivated SubscriptionConditionType = "events-activated"
	// SubscriptionReady label for Knative Subscription readiness in Kyma Subscription
	SubscriptionReady SubscriptionConditionType = "kn-subscription-ready"
)
