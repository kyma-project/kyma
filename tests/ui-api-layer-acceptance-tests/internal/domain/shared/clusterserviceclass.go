package shared

type ClusterServiceClass struct {
	Name                string
	ExternalName        string
	DisplayName         string
	CreationTimestamp   int
	Description         string
	LongDescription     string
	ImageUrl            string
	DocumentationUrl    string
	SupportUrl          string
	ProviderDisplayName string
	Tags                []string
	Activated           bool
	Plans               []ClusterServicePlan
	apiSpec             map[string]interface{}
	asyncApiSpec        map[string]interface{}
	content             map[string]interface{}
}
