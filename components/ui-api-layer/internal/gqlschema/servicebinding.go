package gqlschema

type ServiceBinding struct {
	Name                string
	ServiceInstanceName string
	Environment         string
	SecretName          string
	Status              ServiceBindingStatus
}
