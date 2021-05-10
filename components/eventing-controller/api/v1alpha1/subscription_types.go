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
	ContentMode     *string      `json:"contentMode,omitempty"`
	ExemptHandshake *bool        `json:"exemptHandshake,omitempty"`
	Qos             *string      `json:"qos,omitempty"`
	WebhookAuth     *WebhookAuth `json:"webhookAuth,omitempty"`
}

const (
	ProtocolSettingsContentModeBinary     string = "BINARY"
	ProtocolSettingsContentModeStructured string = "STRUCTURED"
)

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

// BebFilters defines the list of BEB filters
type BebFilters struct {
	Dialect string       `json:"dialect,omitempty"`
	Filters []*BebFilter `json:"filters"`
}

// SubscriptionSpec defines the desired state of Subscription
type SubscriptionSpec struct {
	// ID is the unique identifier of Subscription, read-only.
	ID               string            `json:"id,omitempty"`
	Protocol         string            `json:"protocol"`
	ProtocolSettings *ProtocolSettings `json:"protocolsettings"`
	Sink             string            `json:"sink"`
	Filter           *BebFilters       `json:"filter"`
}

type EmsSubscriptionStatus struct {
	SubscriptionStatus       string `json:"subscriptionStatus,omitempty"`
	SubscriptionStatusReason string `json:"subscriptionStatusReason,omitempty"`
	LastSuccessfulDelivery   string `json:"lastSuccessfulDelivery,omitempty"`
	LastFailedDelivery       string `json:"lastFailedDelivery,omitempty"`
	LastFailedDeliveryReason string `json:"lastFailedDeliveryReason,omitempty"`
}

// +kubebuilder:subresource:status

// SubscriptionStatus defines the observed state of Subscription
type SubscriptionStatus struct {
	Conditions            []Condition           `json:"conditions,omitempty"`
	Ready                 bool                  `json:"ready"`
	Ev2hash               int64                 `json:"ev2hash,omitempty"`
	Emshash               int64                 `json:"emshash,omitempty"`
	ExternalSink          string                `json:"externalSink,omitempty"`
	FailedActivation      string                `json:"failedActivation,omitempty"`
	APIRuleName           string                `json:"apiRuleName,omitempty"`
	EmsSubscriptionStatus EmsSubscriptionStatus `json:"emsSubscriptionStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status

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
