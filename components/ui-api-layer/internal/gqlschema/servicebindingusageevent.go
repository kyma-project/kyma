package gqlschema

type ServiceBindingUsageEvent struct {
	Type         SubscriptionEventType `json:"type"`
	BindingUsage ServiceBindingUsage   `json:"bindingUsage"`
}
