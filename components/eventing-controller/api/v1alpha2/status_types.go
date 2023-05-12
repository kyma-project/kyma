package v1alpha2

type EventType struct {
	// Event type as specified in the Subscription spec.
	OriginalType string `json:"originalType"`
	// Event type after it was cleaned up from backend compatible characters.
	CleanType string `json:"cleanType"`
}

// Backend contains Backend-specific fields.
type Backend struct {
	// EventMesh-specific fields

	// Checksum for the Subscription custom resource.
	// +optional
	Ev2hash int64 `json:"ev2hash,omitempty"`

	// Hash that is used in EventMesh to identify this Subscription.
	// +optional
	Emshash int64 `json:"emshash,omitempty"`

	// Webhook URL used by EventMesh to trigger subscribers.
	// +optional
	ExternalSink string `json:"externalSink,omitempty"`

	// Provides the reason if a Subscription had failed activation in EventMesh.
	// +optional
	FailedActivation string `json:"failedActivation,omitempty"`

	// Name of the APIRule which is used by the Subscription.
	// +optional
	APIRuleName string `json:"apiRuleName,omitempty"`

	// Status of Subscription as reported by EventMesh.
	// +optional
	EmsSubscriptionStatus *EmsSubscriptionStatus `json:"emsSubscriptionStatus,omitempty"`

	// List of event type to consumer name mappings for the NATS backend.
	// +optional
	Types []JetStreamTypes `json:"types,omitempty"`

	// List of mappings from event type to EventMesh compatible types. Used only with EventMesh as the backend.
	// +optional
	EmsTypes []EventMeshTypes `json:"emsTypes,omitempty"`
}

type EmsSubscriptionStatus struct {
	// Status of the Subscription as reported by the backend.
	// +optional
	Status string `json:"status,omitempty"`

	// Reason for the current status.
	// +optional
	StatusReason string `json:"statusReason,omitempty"`

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

type JetStreamTypes struct {
	// Event type that was originally used to subscribe.
	OriginalType string `json:"originalType"`
	// Name of the JetStream consumer created for the Event type.
	ConsumerName string `json:"consumerName,omitempty"`
}

type EventMeshTypes struct {
	// Event type that was originally used to subscribe.
	OriginalType string `json:"originalType"`
	// Event type that is used on the EventMesh backend.
	EventMeshType string `json:"eventMeshType"`
}
