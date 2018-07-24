package gqlschema

type ServiceBindingUsage struct {
	Name               string
	Environment        string
	ServiceBindingName string
	UsedBy             LocalObjectReference
	Status             ServiceBindingUsageStatus
	Parameters         *ServiceBindingUsageParameters
}

type LocalObjectReference struct {
	Kind BindingUsageReferenceType
	Name string
}

type ServiceBindingUsageParameters struct {
	EnvPrefix *EnvPrefix
}

type EnvPrefix struct {
	Name string
}
