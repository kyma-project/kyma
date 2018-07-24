package gqlschema

type LimitRange struct {
	Name   string           `json:"name"`
	Limits []LimitRangeItem `json:"limits"`
}

type LimitRangeItem struct {
	LimitType      LimitType    `json:"limitType"`
	Max            ResourceType `json:"max"`
	Default        ResourceType `json:"default"`
	DefaultRequest ResourceType `json:"defaultRequest"`
}
