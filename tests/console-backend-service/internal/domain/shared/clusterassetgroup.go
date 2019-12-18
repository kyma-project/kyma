package shared

type ClusterAssetGroup struct {
	Name        string           `json:"name"`
	GroupName   string           `json:"groupName"`
	DisplayName string           `json:"displayName"`
	Description string           `json:"description"`
	Assets      []ClusterAsset   `json:"assets"`
	Status      AssetGroupStatus `json:"status"`
}
