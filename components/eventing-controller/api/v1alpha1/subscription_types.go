package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Webhook defines the Webhook called by an active subscription in BEB
type WebhookAuth struct {
	Type         string   `json:"type,omitempty"`
	GrantType    string   `json:"grantType"`
	ClientId     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	TokenUrl     string   `json:"tokenUrl"`
	Scope        []string `json:"scope,omitempty"`
}

// ProtocolSettings defines the CE protocol setting specification implementation
type ProtocolSettings struct {
	ContentMode     string       `json:"contentMode,omitempty"`
	ExemptHandshake bool         `json:"exemptHandshake,omitempty"`
	Qos             string       `json:"qos,omitempty"`
	WebhookAuth     *WebhookAuth `json:"webhookAuth"`
}

// Filter defines the CE filter element
type Filter struct {
	Type     string `json:"type,omitempty"`
	Property string `json:"property"`
	Value    string `json:"value"`
}

// BebFilter defines the BEB filter element as a combination of two CE filter elements
type BebFilter struct {
	EventSource *Filter `json:"eventSource"`
	EventType   *Filter `json:"eventType"`
}

// BebFilters defines the list of Beb filters
type BebFilters struct {
	Dialect string       `json:"dialect,omitempty"`
	Filters []*BebFilter `json:"filters"`
}

// SubscriptionSpec defines the desired state of Subscription
type SubscriptionSpec struct {
	// Id is the unique identifier of Subscription, read-only.
	Id               string            `json:"id,omitempty"`
	Protocol         string            `json:"protocol"`
	ProtocolSettings *ProtocolSettings `json:"protocolsettings"`
	Sink             string            `json:"sink"`
	Filter           *BebFilters       `json:"filter"`
}

// +kubebuilder:subresource:status

// SubscriptionStatus defines the observed state of Subscription
// TODO: it should contain:
// - the status of BEB subscription
// - the status of the exposed Webhook
type SubscriptionStatus struct {
	Ready string `json:"ready"`
}

// +kubebuilder:object:root=true

// Subscription is the Schema for the subscriptions API
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionSpec   `json:"spec,omitempty"`
	Status SubscriptionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubscriptionList contains a list of Subscription
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Subscription{}, &SubscriptionList{})
}
