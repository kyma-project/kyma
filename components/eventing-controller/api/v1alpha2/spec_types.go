package v1alpha2

// ProtocolSettings defines the CE protocol setting specification implementation
type ProtocolSettings struct {
	// ContentMode defines content mode for eventing based on BEB
	// +optional
	ContentMode *string `json:"contentMode,omitempty"`

	// ExemptHandshake defines whether exempt handshake for eventing based on BEB
	// +optional
	ExemptHandshake *bool `json:"exemptHandshake,omitempty"`

	// Qos defines quality of service for eventing based on BEB
	// +optional
	Qos *string `json:"qos,omitempty"`

	// WebhookAuth defines the Webhook called by an active subscription in BEB
	// +optional
	WebhookAuth *WebhookAuth `json:"webhookAuth,omitempty"`
}

type TypeMatching string

const (
	STANDARD TypeMatching = "standard"
	EXACT    TypeMatching = "exact"
)
