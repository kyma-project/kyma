package shared

type ClusterAsset struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
	Type     string                 `json:"type"`
	Files    []File                 `json:"files"`
	Status   AssetStatus            `json:"status"`
}
