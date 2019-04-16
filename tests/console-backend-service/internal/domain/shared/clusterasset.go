package shared

type ClusterAsset struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	Files  []File      `json:"files"`
	Status AssetStatus `json:"status"`
}
