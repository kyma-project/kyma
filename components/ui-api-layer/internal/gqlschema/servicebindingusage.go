package gqlschema

type ServiceBindingUsage struct {
	Name               string
	Environment        string
	ServiceBindingName string
	UsedBy             LocalObjectReference
	Status             ServiceBindingUsageStatus
	Parameters         *ServiceBindingUsageParameters
}
