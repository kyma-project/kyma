package gqlschema

type Asset struct {
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Type       string      `json:"type"`
	Status     AssetStatus `json:"status"`
	Parameters JSON        `json:"parameters"`
}
