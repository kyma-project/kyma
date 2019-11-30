package shared

type ServiceClass struct {
	Name                string
	Namespace           string
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
	Plans               []ServicePlan
	AssetGroup          AssetGroup
}
