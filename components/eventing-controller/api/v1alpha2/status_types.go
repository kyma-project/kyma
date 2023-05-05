package v1alpha2

type EventType struct {
	// OriginalType is the event type specified in the subscription spec
	OriginalType string `json:"originalType"`
	// CleanType is the event type after it was cleaned up from backend compatible characters
	CleanType string `json:"cleanType"`
}

// Backend contains Backend-specific fields.
type Backend struct {
	// EventMesh-specific fields

	// Ev2hash defines the hash for the Subscription custom resource.
	// +optional
	Ev2hash int64 `json:"ev2hash,omitempty"`

	// Emshash defines the hash for the Subscription in EventType.
	// +optional
	Emshash int64 `json:"emshash,omitempty"`

	// ExternalSink defines the webhook URL which is used by EventMesh to trigger subscribers.
	// +optional
	ExternalSink string `json:"externalSink,omitempty"`

	// FailedActivation defines the reason if a Subscription had failed activation in EventMesh.
	// +optional
	FailedActivation string `json:"failedActivation,omitempty"`

	// APIRuleName defines the name of the APIRule which is used by the Subscription.
	// +optional
	APIRuleName string `json:"apiRuleName,omitempty"`

	// EmsSubscriptionStatus defines the status of Subscription in EventMesh.
	// +optional
	EmsSubscriptionStatus *EmsSubscriptionStatus `json:"emsSubscriptionStatus,omitempty"`

	// Types is a list of event type to consumer name mappings for the Nats backend
	// +optional
	Types []JetStreamTypes `json:"types,omitempty"`

	// EmsTypes is a list of mappings between event type and EventMesh compatible types. Only used with EventMesh as backend
	// +optional
	EmsTypes []EventMeshTypes `json:"emsTypes,omitempty"`
}

type EmsSubscriptionStatus struct {
	// Status defines the status of the Subscription.
	// +optional
	Status string `json:"status,omitempty"`

	// StatusReason defines the reason of the status.
	// +optional
	StatusReason string `json:"statusReason,omitempty"`

	// LastSuccessfulDelivery defines the timestamp of the last successful delivery.
	// +optional
	LastSuccessfulDelivery string `json:"lastSuccessfulDelivery,omitempty"`

	// LastFailedDelivery defines the timestamp of the last failed delivery.
	// +optional
	LastFailedDelivery string `json:"lastFailedDelivery,omitempty"`

	// LastFailedDeliveryReason defines the reason of failed delivery.
	// +optional
	LastFailedDeliveryReason string `json:"lastFailedDeliveryReason,omitempty"`
}

type JetStreamTypes struct {
	// OriginalType is the event type originally used to subscribe
	OriginalType string `json:"originalType"`
	// ConsumerName is the name of the Jetstream consumer
	ConsumerName string `json:"consumerName,omitempty"`
}

type EventMeshTypes struct {
	// OriginalType is the event type originally used to subscribe
	OriginalType string `json:"originalType"`
	// EventMeshType is the event type that is used on the event mesh backend
	EventMeshType string `json:"eventMeshType"`
}
