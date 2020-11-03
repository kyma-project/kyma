package types

type Event struct {
	Source string `json:"source,omitempty"`
	Type   string `json:"type,omitempty"`
}

type Events []Event

type Subscription struct {
	Name                     string             `json:"name,omitempty"`
	Events                   Events             `json:"events,omitempty"`
	WebhookUrl               string             `json:"webhookUrl,omitempty"`
	WebhookAuth              *WebhookAuth       `json:"webhookAuth,omitempty"`
	Qos                      Qos                `json:"qos,omitempty"`
	ExemptHandshake          bool               `json:"exemptHandshake,omitempty"`
	ContentMode              string             `json:"contentMode,omitempty"`
	HandshakeStatus          string             `json:"handshakeStatus,omitempty"`
	SubscriptionStatus       SubscriptionStatus `json:"subscriptionStatus,omitempty"`
	SubscriptionStatusReason string             `json:"subscriptionStatusReason,omitempty"`
	LastSuccessfulDelivery   string             `json:"lastSuccessfulDelivery,omitempty"`
	LastFailedDelivery       string             `json:"lastFailedDelivery,omitempty"`
	LastFailedDeliveryReason string             `json:"lastFailedDeliveryReason,omitempty"`
}

type Subscriptions []Subscription

type WebhookAuth struct {
	Type         AuthType  `json:"type,omitempty"`
	User         string    `json:"user,omitempty"`
	Password     string    `json:"password,omitempty"`
	GrantType    GrantType `json:"grantType,omitempty"`
	ClientID     string    `json:"clientId,omitempty"`
	ClientSecret string    `json:"clientSecret,omitempty"`
	TokenURL     string    `json:"tokenUrl,omitempty"`
}

type State struct {
	Action StateAction `json:"action,omitempty"`
}
