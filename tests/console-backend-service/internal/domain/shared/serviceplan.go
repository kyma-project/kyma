package shared

type ServicePlan struct {
	Name                          string
	DisplayName                   string
	ExternalName                  string
	Description                   string
	RelatedServiceClassName       string
	InstanceCreateParameterSchema map[string]interface{}
	BindingCreateParameterSchema  map[string]interface{}
}
