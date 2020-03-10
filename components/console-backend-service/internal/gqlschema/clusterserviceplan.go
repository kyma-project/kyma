package gqlschema

type ClusterServicePlan struct {
	Name                           string  `json:"name"`
	DisplayName                    *string `json:"displayName"`
	ExternalName                   string  `json:"externalName"`
	Description                    string  `json:"description"`
	RelatedClusterServiceClassName string  `json:"relatedClusterServiceClassName"`
	InstanceCreateParameterSchema  *JSON   `json:"instanceCreateParameterSchema"`
	BindingCreateParameterSchema   *JSON   `json:"bindingCreateParameterSchema"`
}
