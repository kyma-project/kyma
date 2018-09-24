package pretty

type Kind int

const (
	ServiceBinding Kind = iota
	ServiceBindings
	ServiceBindingUsage
	ServiceBindingUsages

	ServiceBroker
	ServiceBrokers
	ClusterServiceBroker
	ClusterServiceBrokers

	ServiceClass
	ServiceClasses
	ClusterServiceClass
	ClusterServiceClasses

	ServiceInstance
	ServiceInstances

	ServicePlan
	ServicePlans
	ClusterServicePlan
	ClusterServicePlans

	UsageKind
	UsageKinds
	UsageKindResource
	UsageKindResources
	BindableResources
)

func (k Kind) String() string {
	switch k {
	case ServiceBinding:
		return "Service Binding"
	case ServiceBindings:
		return "Service Bindings"
	case ServiceBindingUsage:
		return "Service Binding Usage"
	case ServiceBindingUsages:
		return "Service Binding Usages"
	case ServiceBroker:
		return "Service Broker"
	case ServiceBrokers:
		return "Service Brokers"
	case ClusterServiceBroker:
		return "Cluster Service Broker"
	case ClusterServiceBrokers:
		return "Cluster Service Brokers"
	case ServiceClass:
		return "Service Class"
	case ServiceClasses:
		return "Service Classes"
	case ClusterServiceClass:
		return "Cluster Service Class"
	case ClusterServiceClasses:
		return "Cluster Service Classes"
	case ServiceInstance:
		return "Service Instance"
	case ServiceInstances:
		return "Service Instances"
	case ServicePlan:
		return "Service Plan"
	case ServicePlans:
		return "Service Plans"
	case ClusterServicePlan:
		return "Cluster Service Plan"
	case ClusterServicePlans:
		return "Cluster Service Plans"
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
