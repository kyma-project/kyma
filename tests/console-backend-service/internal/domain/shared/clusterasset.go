package shared

type ClusterAsset struct {
	Name       string                 `json:"name"`
	Metadata   map[string]interface{} `json:"metadata"`
	Parameters map[string]interface{} `json:"parameters"`
	Type       string                 `json:"type"`
	Files      []File                 `json:"files"`
	Status     AssetStatus            `json:"status"`
}
