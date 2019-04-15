package shared

type ClusterDocsTopic struct {
	Name        string          `json:"name"`
	GroupName   string          `json:"groupName"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Assets      []ClusterAsset  `json:"assets"`
	Status      DocsTopicStatus `json:"status"`
}
