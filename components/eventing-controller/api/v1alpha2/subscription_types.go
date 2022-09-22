package v1alpha2

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TypeMatching string

var Finalizer = GroupVersion.Group

// SubscriptionSpec defines the desired state of Subscription
type SubscriptionSpec struct {
	// ID is the unique identifier of Subscription, read-only
	// +optional
	ID string `json:"id,omitempty"`

	// Sink defines endpoint of the subscriber
	Sink string `json:"sink"`

	// TypeMatching defines the type of matching to be done for the event types
	TypeMatching TypeMatching `json:"typeMatching,omitempty"`

	// Source Defines the source of the event originated from
	Source string `json:"source"`

	// Types defines the list of event names for the topics we need to subscribe for messages
	Types []string `json:"types"`

	// Config defines the configurations that can be applied to the eventing backend
	// +optional
	Config map[string]string `json:"config,omitempty"`
}

// SubscriptionStatus defines the observed state of Subscription
// +kubebuilder:subresource:status
type SubscriptionStatus struct {
	// Conditions defines the status conditions
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// Ready defines the overall readiness status of a subscription
	Ready bool `json:"ready"`

	// Types defines the filter's event types after cleanup for use with the configured backend
	Types []EventType `json:"types"`

	// Backend contains backend specific status which are only applicable to the active backend
	Backend Backend `json:"backend,omitempty"`
}

//+kubebuilder:storageversion
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Clean Event Types",type="string",JSONPath=".status.cleanEventTypes"

// Subscription is the Schema for the subscriptions API
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionSpec   `json:"spec,omitempty"`
	Status SubscriptionStatus `json:"status,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface.
// If the SubscriptionStatus.CleanEventTypes is nil, it will be initialized to an empty slice of stings.
// It is needed because the Kubernetes APIServer will reject requests containing null in the JSON payload.
func (s Subscription) MarshalJSON() ([]byte, error) {
	// Use type alias to copy the subscription without causing an infinite recursion when calling json.Marshal.
	type Alias Subscription
	a := Alias(s)
	if a.Status.Types == nil {
		a.Status.InitializeEventTypes()
	}
	return json.Marshal(a)
}

// InitializeEventTypes initializes the SubscriptionStatus.EventTypes with an empty slice of EventType.
func (s *SubscriptionStatus) InitializeEventTypes() {
	s.Types = []EventType{}
}

//+kubebuilder:object:root=true

// SubscriptionList contains a list of Subscription
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Subscription{}, &SubscriptionList{})
}

// Hub marks this type as a conversion hub.
func (*Subscription) Hub() {}
