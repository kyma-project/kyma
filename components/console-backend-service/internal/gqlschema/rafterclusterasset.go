package gqlschema

type RafterClusterAsset struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Status     RafterAssetStatus `json:"status"`
	Metadata   JSON              `json:"metadata"`
	Parameters JSON              `json:"parameters"`
}
