package automock

// Broker

func NewBrokerGetter() *brokerGetter {
	return new(brokerGetter)
}

func NewBrokerListGetter() *brokerListGetter {
	return new(brokerListGetter)
}

func NewBrokerLister() *brokerLister {
	return new(brokerLister)
}

func NewGQLBrokerConverter() *gqlBrokerConverter {
	return new(gqlBrokerConverter)
}

// Class

func NewClassGetter() *classGetter {
	return new(classGetter)
}

func NewClassInstanceLister() *classInstanceLister {
	return new(classInstanceLister)
}

func NewClassListGetter() *classListGetter {
	return new(classListGetter)
}

func NewGQLClassConverter() *gqlClassConverter {
	return new(gqlClassConverter)
}

// Plan

func NewPlanGetter() *planGetter {
	return new(planGetter)
}

func NewPlanLister() *planLister {
	return new(planLister)
}

func NewGQLPlanConverter() *gqlPlanConverter {
	return new(gqlPlanConverter)
}

// Service Instance

func NewInstanceGetter() *instanceGetter {
	return new(instanceGetter)
}

func NewInstanceLister() *instanceLister {
	return new(instanceLister)
}

// Service Binding

func NewServiceBindingOperations() *serviceBindingOperations {
	return new(serviceBindingOperations)
}

// Service Binding Usage

func NewServiceBindingUsageOperations() *serviceBindingUsageOperations {
	return new(serviceBindingUsageOperations)
}

func NewStatusBindingUsageExtractor() *statusBindingUsageExtractor {
	return new(statusBindingUsageExtractor)
}
