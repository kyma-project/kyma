package shared

type ClusterServicePlan struct {
	Name                           string
	DisplayName                    string
	ExternalName                   string
	Description                    string
	RelatedClusterServiceClassName string
	InstanceCreateParameterSchema  map[string]interface{}
}
