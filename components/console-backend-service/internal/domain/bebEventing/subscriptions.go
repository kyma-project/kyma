package bebEventing

import 
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type EventSubscription struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the desired state of the Trigger.
	Spec SubscriptionSpec `json:"spec,omitempty"`
	// Status represents the current state of the Trigger. This data may be out of
	// date.
	// +optional
	Status SubscriptionStatus `json:"status,omitempty"`
}

type SubscriptionStatus struct {
	ApiRuleName *string
	Conditions *Conditions
	EmsSubscriptionStatus *EmsSubscriptionStatus
	Emshash *int64
	Ev2hash *int64
	ExternalSink *string
	FailedActivation *string
	Ready *bool
}

type Conditions struct {
	Items *[]ConditionItem
}

type ConditionItem struct {
	Item *string
	Message *string
	Reason *string
	Status string
	Type *string
}
type EmsSubscriptionStatus struct {
	LastFailedDelivery *string
	LastFailedDeliveryReason *string
	LastSuccessfulDelivery *string
	SubscriptionStatus *string
	SubscriptionStatusReason *string
}

type SubscriptionSpec struct {
	Filter BebFilters
	Id string
	Protocol string
	ProtocolSettings ProtocolSettings
	Sink string
}
type ProtocolSettings struct {
	ContentMode string
	ExemptHandshake bool
	Qos string
	WebhookAuth Webhook
}
type Webhook struct {
	ClientId string
	ClientSecret string
	GrantType string
	Scope *[]string
	TokenUrl string
	Type *string
}
type BebFilters struct {
	Dialect *string
	Filters []BebFilter
}
type BebFilter struct {
	EventSource Filter
	EventType Filter
}
type Filter struct {
	Property string
	Type *string
	Value string
}