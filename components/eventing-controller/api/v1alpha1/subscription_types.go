package v1alpha1

import (
	"encoding/json"

	"github.com/mitchellh/hashstructure/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

var Finalizer = GroupVersion.Group

// WebhookAuth defines the Webhook called by an active subscription in BEB.
// TODO: Remove it when depreciating code of v1alpha1
type WebhookAuth struct {
	// Defines the authentication type.
	// +optional
	Type string `json:"type,omitempty"`

	// Defines the grant type for OAuth2.
	GrantType string `json:"grantType"`

	// Defines the clientID for OAuth2.
	ClientID string `json:"clientId"`

	// Defines the Client Secret for OAuth2.
	ClientSecret string `json:"clientSecret"`

	// Defines the token URL for OAuth2.
	TokenURL string `json:"tokenUrl"`

	// Defines the scope for OAuth2.
	Scope []string `json:"scope,omitempty"`
}

// ProtocolSettings defines the CE protocol setting specification implementation.
// TODO: Remove it when depreciating code of v1alpha1
type ProtocolSettings struct {
	// Defines the content mode for eventing based on BEB.
	//  The value is either `BINARY`, or `STRUCTURED`.
	// +optional
	ContentMode *string `json:"contentMode,omitempty"`

	// Defines if the exempt handshake for eventing is based on BEB.
	// +optional
	ExemptHandshake *bool `json:"exemptHandshake,omitempty"`

	// Defines the quality of service for eventing based on BEB.
	// +optional
	Qos *string `json:"qos,omitempty"`

	// Defines the Webhook called by an active subscription on BEB.
	// +optional
	WebhookAuth *WebhookAuth `json:"webhookAuth,omitempty"`
}

// TODO: Remove it when depreciating code of v1alpha1
const (
	ProtocolSettingsContentModeBinary     string = "BINARY"
	ProtocolSettingsContentModeStructured string = "STRUCTURED"
)

// Filter defines the CE filter element.
type Filter struct {
	// Defines the type of the filter.
	// +optional
	Type string `json:"type,omitempty"`

	// Defines the property of the filter.
	Property string `json:"property"`

	// Defines the value of the filter.
	Value string `json:"value"`
}

// Defines the BEB filter element as a combination of two CE filter elements.
type EventMeshFilter struct {
	// Defines the source of the CE filter.
	EventSource *Filter `json:"eventSource"`

	// Defines the type of the CE filter.
	EventType *Filter `json:"eventType"`
}

func (bf *EventMeshFilter) hash() (uint64, error) {
	return hashstructure.Hash(bf, hashstructure.FormatV2, nil)
}

// BEBFilters defines the list of BEB filters.
type BEBFilters struct {
	// Contains a `URI-reference` to the CloudEvent filter dialect. See
	// [here](https://github.com/cloudevents/spec/blob/main/subscriptions/spec.md#3241-filter-dialects) for more details.
	// +optional
	Dialect string `json:"dialect,omitempty"`

	Filters []*EventMeshFilter `json:"filters"`
}

// Deduplicate returns a deduplicated copy of BEBFilters.
func (bf *BEBFilters) Deduplicate() (*BEBFilters, error) {
	seen := map[uint64]struct{}{}
	result := &BEBFilters{
		Dialect: bf.Dialect,
	}
	for _, f := range bf.Filters {
		h, err := f.hash()
		if err != nil {
			return nil, err
		}
		if _, exists := seen[h]; !exists {
			result.Filters = append(result.Filters, f)
			seen[h] = struct{}{}
		}
	}
	return result, nil
}

type SubscriptionConfig struct {
	// Defines how many not-ACKed messages can be in flight simultaneously.
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxInFlightMessages int `json:"maxInFlightMessages,omitempty"`
}

// MergeSubsConfigs returns a valid subscription config object based on the provided config,
// complemented with default values, if necessary.
func MergeSubsConfigs(config *SubscriptionConfig, defaults *env.DefaultSubscriptionConfig) *SubscriptionConfig {
	merged := &SubscriptionConfig{
		MaxInFlightMessages: defaults.MaxInFlightMessages,
	}
	if config == nil {
		return merged
	}
	if config.MaxInFlightMessages >= 1 {
		merged.MaxInFlightMessages = config.MaxInFlightMessages
	}
	return merged
}

// SubscriptionSpec defines the desired state of Subscription.
type SubscriptionSpec struct {
	// Unique identifier of the Subscription, read-only.
	// +optional
	ID string `json:"id,omitempty"`

	// Defines the CE protocol specification implementation.
	// +optional
	Protocol string `json:"protocol,omitempty"`

	// Defines the CE protocol settings specification implementation.
	// +optional
	ProtocolSettings *ProtocolSettings `json:"protocolsettings,omitempty"`

	// Kubernetes Service that should be used as a target for the events that match the Subscription.
	// Must exist in the same Namespace as the Subscription.
	Sink string `json:"sink"`

	// Defines which events will be sent to the sink.
	Filter *BEBFilters `json:"filter"`

	// Defines additional configuration for the active backend.
	// +optional
	Config *SubscriptionConfig `json:"config,omitempty"`
}

type EmsSubscriptionStatus struct {
	// Status of the Subscription as reported by EventMesh.
	// +optional
	SubscriptionStatus string `json:"subscriptionStatus,omitempty"`

	// Reason for the current status.
	// +optional
	SubscriptionStatusReason string `json:"subscriptionStatusReason,omitempty"`

	// Timestamp of the last successful delivery.
	// +optional
	LastSuccessfulDelivery string `json:"lastSuccessfulDelivery,omitempty"`

	// Timestamp of the last failed delivery.
	// +optional
	LastFailedDelivery string `json:"lastFailedDelivery,omitempty"`

	// Reason for the last failed delivery.
	// +optional
	LastFailedDeliveryReason string `json:"lastFailedDeliveryReason,omitempty"`
}

// SubscriptionStatus defines the observed state of the Subscription.
type SubscriptionStatus struct {
	// Current state of the Subscription.
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// Overall readiness of the Subscription.
	Ready bool `json:"ready"`

	// CleanEventTypes defines the filter's event types after cleanup to use it with the configured backend.
	CleanEventTypes []string `json:"cleanEventTypes"`

	// Defines the checksum for the Subscription custom resource.
	// +optional
	Ev2hash int64 `json:"ev2hash,omitempty"`

	// Defines the checksum for the Subscription in EventMesh.
	// +optional
	Emshash int64 `json:"emshash,omitempty"`

	// Defines the webhook URL which is used by EventMesh to trigger subscribers.
	// +optional
	ExternalSink string `json:"externalSink,omitempty"`

	// Defines the reason if a Subscription failed activation in EventMesh.
	// +optional
	FailedActivation string `json:"failedActivation,omitempty"`

	// Defines the name of the APIRule which is used by the Subscription.
	// +optional
	APIRuleName string `json:"apiRuleName,omitempty"`

	// Defines the status of the Subscription in EventMesh.
	// +optional
	EmsSubscriptionStatus *EmsSubscriptionStatus `json:"emsSubscriptionStatus,omitempty"`

	// Defines the configurations that have been applied to the eventing backend when creating this Subscription.
	// +optional
	Config *SubscriptionConfig `json:"config,omitempty"`
}

// Subscription is the Schema for the subscriptions API.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:deprecatedversion:warning=The v1alpha1 API version is deprecated as of Kyma 2.14.X.
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Clean Event Types",type="string",JSONPath=".status.cleanEventTypes"
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
	if a.Status.CleanEventTypes == nil {
		a.Status.InitializeCleanEventTypes()
	}
	return json.Marshal(a)
}

// SubscriptionList contains a list of Subscription.
// +kubebuilder:object:root=true
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}

// InitializeCleanEventTypes initializes the SubscriptionStatus.CleanEventTypes with an empty slice of strings.
func (s *SubscriptionStatus) InitializeCleanEventTypes() {
	s.CleanEventTypes = []string{}
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&Subscription{}, &SubscriptionList{})
}
