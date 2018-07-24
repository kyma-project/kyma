package gqlschema

type ResourceQuota struct {
	Name     string
	Pods     *string
	Limits   ResourceValues
	Requests ResourceValues
}
