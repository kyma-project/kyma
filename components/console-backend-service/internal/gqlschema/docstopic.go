package gqlschema

type DocsTopic struct {
	Name        string          `json:"name"`
	Namespace   string          `json:"namespace"`
	GroupName   string          `json:"groupName"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Status      DocsTopicStatus `json:"status"`
}
