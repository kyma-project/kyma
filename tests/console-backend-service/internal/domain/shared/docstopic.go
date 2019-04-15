package shared

type DocsTopic struct {
	Name        string          `json:"name"`
	Namespace   string          `json:"namespace"`
	GroupName   string          `json:"groupName"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	Assets      []Asset         `json:"assets"`
	Status      DocsTopicStatus `json:"status"`
}

type DocsTopicStatus struct {
	Phase   DocsTopicPhaseType `json:"phase"`
	Reason  string             `json:"reason"`
	Message string             `json:"message"`
}

type DocsTopicPhaseType string

const (
	DocsTopicPhaseTypeReady DocsTopicPhaseType = "READY"
)
