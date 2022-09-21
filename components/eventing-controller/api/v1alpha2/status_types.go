package v1alpha2

type EventType struct {
	OriginalType string `json:"originalType"`
	CleanType    string `json:"cleanType"`
}

// Backend contains Backend-specific fields
type Backend struct {
	// BEB-specific fields

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

	// NATS-specific fields

	// +optional
	Types []JetStreamTypes `json:"types,omitempty"`
}

type EmsSubscriptionStatus struct {
	// Status defines the status of the Subscription
	// +optional
	Status string `json:"status,omitempty"`

	// StatusReason defines the reason of the status
	// +optional
	StatusReason string `json:"statusReason,omitempty"`

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

type JetStreamTypes struct {
	OriginalType string `json:"originalType"`
	ConsumerName string `json:"consumerName,omitempty"`
}
