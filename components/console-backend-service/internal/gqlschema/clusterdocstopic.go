package gqlschema

type ClusterDocsTopic struct {
	Name        string          `json:"name"`
	GroupName   string          `json:"groupName"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Status      DocsTopicStatus `json:"status"`
}
