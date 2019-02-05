package gqlschema

type ServiceBinding struct {
	Name                string
	ServiceInstanceName string
	Namespace           string
	SecretName          string
	Status              ServiceBindingStatus
	Parameters          JSON
}
