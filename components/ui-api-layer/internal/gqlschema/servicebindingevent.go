package gqlschema

type ServiceBindingEvent struct {
	Type    SubscriptionEventType `json:"type"`
	Binding ServiceBinding        `json:"binding"`
}
