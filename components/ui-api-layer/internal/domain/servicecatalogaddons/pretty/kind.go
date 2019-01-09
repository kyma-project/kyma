package pretty

type Kind int

const (
	ServiceBindingUsage Kind = iota
	ServiceBindingUsages

	UsageKind
	UsageKinds
	UsageKindResource
	UsageKindResources
	BindableResources
)

func (k Kind) String() string {
	switch k {
	case ServiceBindingUsage:
		return "Service Binding Usage"
	case ServiceBindingUsages:
		return "Service Binding Usages"
	case UsageKind:
		return "Usage Kind"
	case UsageKinds:
		return "Usage Kinds"
	case UsageKindResource:
		return "Usage Kind Resource"
	case UsageKindResources:
		return "Usage Kind Resources"
	case BindableResources:
		return "Bindable Resources"
	default:
		return ""
	}
}
