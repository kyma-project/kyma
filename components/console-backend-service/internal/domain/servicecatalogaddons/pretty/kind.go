package pretty

type Kind int

const (
	ServiceBindingUsage Kind = iota
	ServiceBindingUsages

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
