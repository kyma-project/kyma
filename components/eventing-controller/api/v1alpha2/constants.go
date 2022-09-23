package v1alpha2

const (
	TypeMatchingStandard TypeMatching = "standard"
	TypeMatchingExact    TypeMatching = "exact"

	// config fields
	MaxInFlightMessages = "maxInFlightMessages"

	// protocol settings
	Protocol                        = "protocol"
	ProtocolSettingsContentMode     = "contentMode"
	ProtocolSettingsExemptHandshake = "exemptHandshake"
	ProtocolSettingsQos             = "qos"

	// webhook auth fields
	WebhookAuthType         = "type"
	WebhookAuthGrantType    = "grantType"
	WebhookAuthClientID     = "clientId"
	WebhookAuthClientSecret = "clientSecret"
	WebhookAuthTokenURL     = "tokenUrl"
	WebhookAuthScope        = "scope"
)
