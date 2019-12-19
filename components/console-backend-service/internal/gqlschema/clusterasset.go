package gqlschema

type ClusterAsset struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Status     AssetStatus `json:"status"`
	Parameters JSON        `json:"parameters"`
}
