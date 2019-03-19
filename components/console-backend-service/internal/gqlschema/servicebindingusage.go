package gqlschema

type ServiceBindingUsage struct {
	Name               string
	Namespace          string
	ServiceBindingName string
	UsedBy             LocalObjectReference
	Status             ServiceBindingUsageStatus
	Parameters         *ServiceBindingUsageParameters
}
