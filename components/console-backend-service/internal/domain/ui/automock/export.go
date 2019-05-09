package automock

func NewBackendModuleConverter() *gqlBackendModuleConverter {
	return new(gqlBackendModuleConverter)
}

func NewBackendModuleService() *backendModuleLister {
	return new(backendModuleLister)
}

func NewMicrofrontendConverter() *gqlMicrofrontendConverter {
	return new(gqlMicrofrontendConverter)
}

func NewMicrofrontendService() *microfrontendLister {
	return new(microfrontendLister)
}

func NewClusterMicrofrontendConverter() *gqlClusterMicrofrontendConverter {
	return new(gqlClusterMicrofrontendConverter)
}

func NewClusterMicrofrontendService() *clusterMicrofrontendLister {
	return new(clusterMicrofrontendLister)
}
