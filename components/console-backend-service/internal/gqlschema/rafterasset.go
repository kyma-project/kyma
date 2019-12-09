package gqlschema

type RafterAsset struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	Status     RafterAssetStatus `json:"status"`
	Metadata   JSON              `json:"metadata"`
	Parameters JSON              `json:"parameters"`
}
