package gqlschema

type AssetGroup struct {
	Name        string          `json:"name"`
	Namespace   string          `json:"namespace"`
	GroupName   string          `json:"groupName"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Status      AssetGroupStatus `json:"status"`
}
