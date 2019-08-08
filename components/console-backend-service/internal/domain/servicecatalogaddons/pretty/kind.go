package pretty

type Kind int

const (
	ServiceBindingUsage Kind = iota
	ServiceBindingUsages

	AddonsConfiguration
	AddonsConfigurations
	ClusterAddonsConfiguration
	ClusterAddonsConfigurations

	UsageKind
	UsageKinds
	BindableResources
)

func (k Kind) String() string {
	switch k {
	case ServiceBindingUsage:
		return "Service Binding Usage"
	case ServiceBindingUsages:
		return "Service Binding Usages"
	case ClusterAddonsConfiguration:
		return "Cluster Addons Configuration"
	case ClusterAddonsConfigurations:
		return "Cluster Addons Configurations"
	case AddonsConfiguration:
		return "Addons Configuration"
	case AddonsConfigurations:
		return "Addons Configurations"
	case UsageKind:
		return "Usage Kind"
	case UsageKinds:
		return "Usage Kinds"
	case BindableResources:
		return "Bindable Resources"
	default:
		return ""
	}
}
