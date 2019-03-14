package shared

type ServiceBinding struct {
	Name                string
	ServiceInstanceName string
	Namespace           string
	Secret              Secret
	Status              ServiceBindingStatus
}

type Secret struct {
	Name      string
	Namespace string
	Data      map[string]interface{}
}

type ServiceBindingStatus struct {
	Type ServiceBindingStatusType
}

type ServiceBindingStatusType string

const (
	ServiceBindingStatusTypeReady ServiceBindingStatusType = "READY"
)
