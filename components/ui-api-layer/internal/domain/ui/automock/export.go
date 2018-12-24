package automock

func NewBackendModuleConverter() *gqlBackendModuleConverter {
	return new(gqlBackendModuleConverter )
}

func NewBackendModuleService() *backendModuleLister {
	return new(backendModuleLister)
}
