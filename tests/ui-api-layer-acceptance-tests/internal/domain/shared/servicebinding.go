package shared

type ServiceBinding struct {
	Name                string
	ServiceInstanceName string
	Environment         string
	Secret              Secret
	Status              ServiceBindingStatus
}

type Secret struct {
	Name        string
	Environment string
	Data        map[string]interface{}
}

type ServiceBindingStatus struct {
	Type ServiceBindingStatusType
}

type ServiceBindingStatusType string

const (
	ServiceBindingStatusTypeReady ServiceBindingStatusType = "READY"
)
