package gqlschema

type ResourceQuotasStatus struct {
	Exceeded       bool            `json:"exceeded"`
	ExceededQuotas []ExceededQuota `json:"exceededQuotas"`
}

type ExceededQuota struct {
	Name              string              `json:"name"`
	ResourcesRequests []ResourcesRequests `json:"resourcesRequests"`
}
