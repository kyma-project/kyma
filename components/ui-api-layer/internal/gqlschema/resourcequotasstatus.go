package gqlschema

type ResourceQuotasStatus struct {
	Exceeded       bool            `json:"exceeded"`
	ExceededQuotas []ExceededQuota `json:"exceededQuotas"`
}

type ExceededQuota struct {
	QuotaName         string   `json:"name"`
	ResourceName      string   `json:"resourceName"`
	AffectedResources []string `json:"affectedResources"`
}
