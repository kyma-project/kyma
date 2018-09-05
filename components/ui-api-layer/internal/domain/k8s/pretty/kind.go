package pretty

type Kind int

const (
	Deployment Kind = iota
	Deployments
	Environment
	Environments
	LimitRange
	LimitRanges
	ResourceQuota
	ResourceQuotas
	ResourceQuotaStatus
	ResourceQuotaStatuses
	Secret
)

func (k Kind) String() string {
	switch k {
	case Deployment:
		return "Deployment"
	case Deployments:
		return "Deployments"
	case Environment:
		return "Environment"
	case Environments:
		return "Environments"
	case LimitRange:
		return "Limit Range"
	case LimitRanges:
		return "Limit Ranges"
	case ResourceQuota:
		return "Resource Quota"
	case ResourceQuotas:
		return "Resource Quotas"
	case ResourceQuotaStatus:
		return "Resource Quota Status"
	case ResourceQuotaStatuses:
		return "Resource Quota Statuses"
	case Secret:
		return "Secret"
	default:
		return ""
	}
}
