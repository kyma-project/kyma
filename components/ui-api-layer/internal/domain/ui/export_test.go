package ui

import "k8s.io/client-go/tools/cache"

func NewBackendModuleService(informer cache.SharedIndexInformer) *backendModuleService {
	return newBackendModuleService(informer)
}

func NewBackendModuleResolver(backendModuleLister backendModuleLister) *backendModuleResolver {
	return newBackendModuleResolver(backendModuleLister)
}

func NewMockBackendModuleConverter() *mockGqlBackendModuleConverter {
	return new(mockGqlBackendModuleConverter )
}

func NewMockBackendModuleService() *mockBackendModuleLister {
	return new(mockBackendModuleLister)
}

func (r *backendModuleResolver) SetInstanceConverter(converter gqlBackendModuleConverter) {
	r.backendModuleConverter = converter
}