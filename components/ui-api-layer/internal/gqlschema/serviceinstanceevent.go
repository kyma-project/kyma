package gqlschema

type ServiceInstanceEvent struct {
	Type     SubscriptionEventType `json:"type"`
	Instance ServiceInstance       `json:"instance"`
}
