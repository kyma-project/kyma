package pretty

type Kind int

const (
	Deployment Kind = iota
	Deployments
	Namespace
	Namespaces
	LimitRange
	LimitRanges
	Pod
	Pods
	ReplicaSet
	ReplicaSets
	ConfigMap
	ConfigMaps
	Resource
	Resources
	StatefulSets
	ResourceQuota
	ResourceQuotas
	ResourceQuotaStatus
	ResourceQuotaStatuses
	Secret
	Secrets
	Service
	Services
	SelfSubjectRules
)

func (k Kind) String() string {
	switch k {
	case Deployment:
		return "Deployment"
	case Deployments:
		return "Deployments"
	case Namespace:
		return "Namespace"
	case Namespaces:
		return "Namespaces"
	case LimitRange:
		return "Limit Range"
	case LimitRanges:
		return "Limit Ranges"
	case Pod:
		return "Pod"
	case Pods:
		return "Pods"
	case ReplicaSet:
		return "Replica Set"
	case ReplicaSets:
		return "Replica Sets"
	case ConfigMap:
		return "ConfigMap"
	case ConfigMaps:
		return "ConfigMaps"
	case Resource:
		return "Resource"
	case Resources:
		return "Resources"
	case StatefulSets:
		return "Stateful Sets"
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
	case Secrets:
		return "Secrets"
	case Service:
		return "Service"
	case Services:
		return "Services"
	case SelfSubjectRules:
		return "SelfSubjectRules"
	default:
		return ""
	}
}
