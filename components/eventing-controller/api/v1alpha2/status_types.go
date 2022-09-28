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
	// +kubebuilder:validation:Minimum=1
	MaxInFlightMessages int `json:"maxInFlightMessages,omitempty"`

	// +optional
	Types []JetStreamTypes `json:"types,omitempty"`

	// +optional
	EmsTypes []EventMeshTypes `json:"emsTypes,omitempty"`
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

// WebhookAuth defines the Webhook called by an active subscription in BEB
type WebhookAuth struct {
	// Type defines type of authentication
	// +optional
	Type string `json:"type,omitempty"`

	// GrantType defines grant type for OAuth2
	GrantType string `json:"grantType"`

	// ClientID defines clientID for OAuth2
	ClientID string `json:"clientId"`

	// ClientSecret defines client secret for OAuth2
	ClientSecret string `json:"clientSecret"`

	// TokenURL defines token URL for OAuth2
	TokenURL string `json:"tokenUrl"`

	// Scope defines scope for OAuth2
	Scope []string `json:"scope,omitempty"`
}

type JetStreamTypes struct {
	OriginalType string `json:"originalType"`
	ConsumerName string `json:"consumerName"`
}

type EventMeshTypes struct {
	OriginalType  string `json:"originalType"`
	EventMeshType string `json:"eventMeshType"`
}
