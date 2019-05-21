package automock

func NewBackendModuleConverter() *gqlBackendModuleConverter {
	return new(gqlBackendModuleConverter)
}

func NewBackendModuleService() *backendModuleLister {
	return new(backendModuleLister)
}

func NewMicroFrontendConverter() *gqlMicroFrontendConverter {
	return new(gqlMicroFrontendConverter)
}

func NewMicroFrontendService() *microFrontendLister {
	return new(microFrontendLister)
}

func NewClusterMicroFrontendConverter() *gqlClusterMicroFrontendConverter {
	return new(gqlClusterMicroFrontendConverter)
}

func NewClusterMicroFrontendService() *clusterMicroFrontendLister {
	return new(clusterMicroFrontendLister)
}
