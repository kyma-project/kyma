package automock

// ClusterServiceBroker

func NewClusterServiceBrokerService() *clusterServiceBrokerSvc {
	return new(clusterServiceBrokerSvc)
}

func NewGQLClusterServiceBrokerConverter() *gqlClusterServiceBrokerConverter {
	return new(gqlClusterServiceBrokerConverter)
}

// ServiceBroker
func NewServiceBrokerService() *serviceBrokerSvc {
	return new(serviceBrokerSvc)
}

func NewGQLServiceBrokerConverter() *gqlServiceBrokerConverter {
	return new(gqlServiceBrokerConverter)
}

// ServiceClass

func NewInstanceListerByServiceClass() *instanceListerByServiceClass {
	return new(instanceListerByServiceClass)
}

func NewServiceClassListGetter() *serviceClassListGetter {
	return new(serviceClassListGetter)
}

func NewGQLServiceClassConverter() *gqlServiceClassConverter {
	return new(gqlServiceClassConverter)
}

func NewServiceClassGetter() *serviceClassGetter {
	return new(serviceClassGetter)
}

// ClusterServiceClass

func NewInstanceListerByClusterServiceClass() *instanceListerByClusterServiceClass {
	return new(instanceListerByClusterServiceClass)
}

func NewClusterServiceClassListGetter() *clusterServiceClassListGetter {
	return new(clusterServiceClassListGetter)
}

func NewGQLClusterServiceClassConverter() *gqlClusterServiceClassConverter {
	return new(gqlClusterServiceClassConverter)
}

// ServicePlan

func NewServicePlanGetter() *servicePlanGetter {
	return new(servicePlanGetter)
}

func NewServicePlanLister() *servicePlanLister {
	return new(servicePlanLister)
}

func NewGQLServicePlanConverter() *gqlServicePlanConverter {
	return new(gqlServicePlanConverter)
}

// ClusterServicePlan

func NewClusterServicePlanGetter() *clusterServicePlanGetter {
	return new(clusterServicePlanGetter)
}

func NewClusterServicePlanLister() *clusterServicePlanLister {
	return new(clusterServicePlanLister)
}

func NewGQLClusterServicePlanConverter() *gqlClusterServicePlanConverter {
	return new(gqlClusterServicePlanConverter)
}

// Service Instance

func NewServiceInstanceLister() *serviceInstanceLister {
	return new(serviceInstanceLister)
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

// Usage Kind

func NewUsageKindServices() *usageKindServices {
	return new(usageKindServices)
}
