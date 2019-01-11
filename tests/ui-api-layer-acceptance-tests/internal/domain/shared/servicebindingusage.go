package shared

type ServiceBindingUsage struct {
	Name           string
	Environment    string
	ServiceBinding ServiceBinding
	UsedBy         LocalObjectReference
	Status         ServiceBindingUsageStatus
}

type LocalObjectReference struct {
	Kind string
	Name string
}

type ServiceBindingUsageStatus struct {
	Type ServiceBindingUsageStatusType
}

type ServiceBindingUsageStatusType string

const (
	ServiceBindingUsageStatusTypeUnknown ServiceBindingUsageStatusType = "UNKNOWN"
)
