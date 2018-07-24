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
	IncludeTopicHeader            bool   `json:"include_topic_header"`
	MaxInflight                   int    `json:"max_inflight"`
	PushRequestTimeoutMS          int64  `json:"push_request_timeout_ms"`
	EventType                     string `json:"event_type"`
	EventTypeVersion              string `json:"event_type_version"`
	Source                        Source `json:"source"`
}

type Source struct {
	SourceNamespace   string `json:"source_namespace"`
	SourceType        string `json:"source_type"`
	SourceEnvironment string `json:"source_environment"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//SubscriptionList
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Subscription `json:"items"`
}

type SubscriptionStatus struct {
	Conditions []SubscriptionCondition `json:"conditions"`
}

type SubscriptionCondition struct {
	Type               SubscriptionConditionType `json:"type"`
	Status             ConditionStatus           `json:"status"`
	LastTransitionTime metav1.Time               `json:"last_transition_time"`
	Reason             string                    `json:"reason"`
	Message            string                    `json:"message"`
}
type SubscriptionConditionType string

const (
	EventsActivated SubscriptionConditionType = "events-activated"
)

type ConditionStatus string

const (
	ConditionTrue  ConditionStatus = "True"
	ConditionFalse ConditionStatus = "False"
)
