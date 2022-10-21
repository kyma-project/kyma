package v1alpha2

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"

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

// Subscription is the Schema for the subscriptions API.
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

// GetMaxInFlightMessages tries to convert the string-type maxInFlight to the integer
// and returns the error in case the conversion is not successful.
func (s *Subscription) GetMaxInFlightMessages() (int, error) {
	if s.Spec.Config == nil {
		return env.DefaultMaxInFlight, nil
	}
	if _, ok := s.Spec.Config[MaxInFlightMessages]; !ok {
		return env.DefaultMaxInFlight, nil
	}
	val, err := strconv.Atoi(s.Spec.Config[MaxInFlightMessages])
	if err != nil {
		return -1, err
	}
	return val, nil
}

// InitializeEventTypes initializes the SubscriptionStatus.Types with an empty slice of EventType.
func (s *SubscriptionStatus) InitializeEventTypes() {
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

func (s *Subscription) GetEventMeshProtocol() string {
	if protocol, ok := s.Spec.Config[Protocol]; ok {
		return protocol
	}
	return ""
}

func (s *Subscription) GetEventMeshProtocolSettings() *ProtocolSettings {
	protocolSettings := &ProtocolSettings{}

	if currentMode, ok := s.Spec.Config[ProtocolSettingsContentMode]; ok {
		protocolSettings.ContentMode = &currentMode
	}
	if qos, ok := s.Spec.Config[ProtocolSettingsQos]; ok {
		protocolSettings.Qos = &qos
	}
	if exemptHandshake, ok := s.Spec.Config[ProtocolSettingsExemptHandshake]; ok {
		handshake, err := strconv.ParseBool(exemptHandshake)
		if err != nil {
			handshake = true
		}
		protocolSettings.ExemptHandshake = &handshake
	}
	if authType, ok := s.Spec.Config[WebhookAuthType]; ok {
		protocolSettings.WebhookAuth = &WebhookAuth{}
		protocolSettings.WebhookAuth.Type = authType
	}
	if grantType, ok := s.Spec.Config[WebhookAuthGrantType]; ok {
		protocolSettings.WebhookAuth.GrantType = grantType
	}
	if clientID, ok := s.Spec.Config[WebhookAuthClientID]; ok {
		protocolSettings.WebhookAuth.ClientID = clientID
	}
	if secret, ok := s.Spec.Config[WebhookAuthClientSecret]; ok {
		protocolSettings.WebhookAuth.ClientSecret = secret
	}
	if token, ok := s.Spec.Config[WebhookAuthTokenURL]; ok {
		protocolSettings.WebhookAuth.TokenURL = token
	}
	if scope, ok := s.Spec.Config[WebhookAuthScope]; ok {
		protocolSettings.WebhookAuth.Scope = strings.Split(scope, ",")
	}

	return protocolSettings
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
