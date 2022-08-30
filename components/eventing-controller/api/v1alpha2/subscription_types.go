package v1alpha2

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TypeMatching string

const (
	STANDARD TypeMatching = "standard"
	EXACT    TypeMatching = "exact"
)

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

	TypeMatching TypeMatching `json:"typeMatching,omitempty"`

	Source string `json:"source,omitempty"`

	Types []string `json:"types,omitempty"`

	// Config defines the configurations that can be applied to the eventing backend when creating this subscription
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
	Types []string `json:"types"`

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
	EmsSubscriptionStatus *EmsSubscriptionStatus `json:"emsSubscriptionStatus,omitempty"`
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
		a.Status.InitializeCleanEventTypes()
	}
	return json.Marshal(a)
}

// InitializeCleanEventTypes initializes the SubscriptionStatus.CleanEventTypes with an empty slice of strings.
func (s *SubscriptionStatus) InitializeCleanEventTypes() {
	s.Types = []string{}
}

//+kubebuilder:object:root=true

// SubscriptionList contains a list of Subscription
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Subscription{}, &SubscriptionList{})
}

// Hub marks this type as a conversion hub.
func (*Subscription) Hub() {}
