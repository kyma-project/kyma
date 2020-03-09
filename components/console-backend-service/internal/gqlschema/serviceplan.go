package gqlschema

type ServicePlan struct {
	Name                          string  `json:"name"`
	Namespace                     string  `json:"namespace"`
	DisplayName                   *string `json:"displayName"`
	ExternalName                  string  `json:"externalName"`
	Description                   string  `json:"description"`
	RelatedServiceClassName       string  `json:"relatedServiceClassName"`
	InstanceCreateParameterSchema *JSON   `json:"instanceCreateParameterSchema"`
	BindingCreateParameterSchema  *JSON   `json:"bindingCreateParameterSchema"`
}
