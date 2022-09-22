package v1alpha2

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Finalizer = GroupVersion.Group

const ProtocolSettingsContentModeBinary string = "BINARY"

// SubscriptionSpec defines the desired state of Subscription
type SubscriptionSpec struct {
	// ID is the unique identifier of Subscription, read-only
	// +optional
	ID string `json:"id,omitempty"`

	// todo check if we still need it or move it to config
	// Protocol defines the CE protocol specification implementation
	// +optional
	Protocol string `json:"protocol,omitempty"`

	// todo check if we still need it or move it to config
	// ProtocolSettings defines the CE protocol setting specification implementation
	// +optional
	ProtocolSettings *ProtocolSettings `json:"protocolsettings,omitempty"`

	// Sink defines endpoint of the subscriber
	Sink string `json:"sink"`

	// TypeMatching defines the type of matching to be done for the event types
	TypeMatching TypeMatching `json:"typeMatching,omitempty"`

	// Source Defines the source of the event originated from
	Source string `json:"source"`

	// Types defines the list of event names for the topics we need to subscribe for messages
	Types []string `json:"types,omitempty"`

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
		a.Status.InitializeCleanEventTypes()
	}
	return json.Marshal(a)
}

// InitializeCleanEventTypes initializes the SubscriptionStatus.CleanEventTypes with an empty slice of strings.
func (s *SubscriptionStatus) InitializeCleanEventTypes() {
	s.Types = []EventType{}
}

// GetUniqueTypes returns the de-duplicated types from subscription spec.
func (s *Subscription) GetUniqueTypes() []string {
	result := make([]string, 0, len(s.Spec.Types))
	for _, t := range s.Spec.Types {
		if !utils.ContainsString(result, t) {
			result = append(result, t)
		}
	}

	return result
}

func (s *Subscription) DuplicateWithStatusDefaults() *Subscription {
	desiredSub := s.DeepCopy()
	desiredSub.Status = SubscriptionStatus{}
	return desiredSub
}

func (s *Subscription) ToUnstructuredSub() (*unstructured.Unstructured, error) {
	object, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(&s)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
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
