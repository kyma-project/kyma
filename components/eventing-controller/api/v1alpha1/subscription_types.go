package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebhookAuth defines the Webhook called by an active subscription in BEB
type WebhookAuth struct {
	// +optional
	Type string `json:"type,omitempty"`

	GrantType string `json:"grantType"`

	ClientId string `json:"clientId"`

	ClientSecret string `json:"clientSecret"`

	TokenUrl string `json:"tokenUrl"`

	Scope []string `json:"scope,omitempty"`
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
	// Type defines the type of the filter
	// +optional
	Type string `json:"type,omitempty"`

	// Property defines the property of the filter
	Property string `json:"property"`

	// Value defines the value of the filter
	Value string `json:"value"`
}

// BebFilter defines the BEB filter element as a combination of two CE filter elements
type BebFilter struct {
	// EventSource defines the source of CE filter
	EventSource *Filter `json:"eventSource"`

	// EventType defines the type of CE filter
	EventType *Filter `json:"eventType"`
}

// BebFilters defines the list of BEB filters
type BebFilters struct {
	// +optional
	Dialect string `json:"dialect,omitempty"`

	Filters []*BebFilter `json:"filters"`
}

// SubscriptionSpec defines the desired state of Subscription
type SubscriptionSpec struct {
	// ID is the unique identifier of Subscription, read-only.
	// +optional
	ID string `json:"id,omitempty"`

	// Protocol defines the CE protocol specification implementation
	// +optional
	Protocol string `json:"protocol,omitempty"`

	// ProtocolSettings defines the CE protocol setting specification implementation
	// +optional
	ProtocolSettings *ProtocolSettings `json:"protocolsettings,omitempty"`

	// Sink defines endpoint of the subscriber
	Sink string `json:"sink"`

	// Filter defines the list of filters
	Filter *BebFilters `json:"filter"`
}

type EmsSubscriptionStatus struct {
	// SubscriptionStatus defines the status of the Subscription
	// +optional
	SubscriptionStatus string `json:"subscriptionStatus,omitempty"`

	// SubscriptionStatusReason defines the reason of the status
	// +optional
	SubscriptionStatusReason string `json:"subscriptionStatusReason,omitempty"`

	// LastSuccessfulDelivery defines the timestamp of the last successful delivery
	// +optional
	LastSuccessfulDelivery string `json:"lastSuccessfulDelivery,omitempty"`

	// LastFailedDelivery defines the timestamp of the last failed delivery
	// +optional
	LastFailedDelivery string `json:"lastFailedDelivery,omitempty"`

	// LastFailedDeliveryReason defines the reason of failed delivery
	// +optional
	LastFailedDeliveryReason string `json:"lastFailedDeliveryReason,omitempty"`
}

// +kubebuilder:subresource:status
// SubscriptionStatus defines the observed state of Subscription
type SubscriptionStatus struct {
	// Conditions defines the status conditions
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// Ready defines the overall readiness status of a subscription
	Ready bool `json:"ready"`

	// Ev2hash defines the hash for the Subscription custom resource
	// +optional
	Ev2hash int64 `json:"ev2hash,omitempty"`

	// Emshash defines the hash for the Subscription in BEB
	// +optional
	Emshash int64 `json:"emshash,omitempty"`

	// ExternalSink defines the webhook URL which is used by BEB to trigger subscribers
	// +optional
	ExternalSink string `json:"externalSink,omitempty"`

	// FailedActivation defines the reason if a Subscription had failed activation in BEB
	// +optional
	FailedActivation string `json:"failedActivation,omitempty"`

	// APIRuleName defines the name of the APIRule which is used by the Subscription
	// +optional
	APIRuleName string `json:"apiRuleName,omitempty"`

	// EmsSubscriptionStatus defines the status of Subscription in BEB
	// +optional
	EmsSubscriptionStatus EmsSubscriptionStatus `json:"emsSubscriptionStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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
