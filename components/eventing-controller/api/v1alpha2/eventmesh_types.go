package v1alpha2

const (
	ProtocolSettingsContentModeBinary     string = "BINARY"
	ProtocolSettingsContentModeStructured string = "STRUCTURED"
)

// WebhookAuth defines the Webhook called by an active subscription in BEB.
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

// ProtocolSettings defines the CE protocol setting specification implementation.
type ProtocolSettings struct {
	// ContentMode defines content mode for eventing based on BEB.
	// +optional
	ContentMode *string `json:"contentMode,omitempty"`

	// ExemptHandshake defines whether exempt handshake for eventing based on BEB.
	// +optional
	ExemptHandshake *bool `json:"exemptHandshake,omitempty"`

	// Qos defines quality of service for eventing based on BEB.
	// +optional
	Qos *string `json:"qos,omitempty"`

	// WebhookAuth defines the Webhook called by an active subscription in BEB.
	// +optional
	WebhookAuth *WebhookAuth `json:"webhookAuth,omitempty"`
}
