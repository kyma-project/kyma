package gqlschema

type ClusterAssetGroup struct {
	Name        string           `json:"name"`
	GroupName   string           `json:"groupName"`
	DisplayName string           `json:"displayName"`
	Description string           `json:"description"`
	Status      AssetGroupStatus `json:"status"`
}
