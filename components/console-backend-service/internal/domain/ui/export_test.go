package ui

import "k8s.io/client-go/tools/cache"

func NewBackendModuleService(informer cache.SharedIndexInformer) *backendModuleService {
	return newBackendModuleService(informer)
}

func NewBackendModuleResolver(backendModuleLister backendModuleLister) *backendModuleResolver {
	return newBackendModuleResolver(backendModuleLister)
}

func (r *backendModuleResolver) SetInstanceConverter(converter gqlBackendModuleConverter) {
	r.backendModuleConverter = converter
}

func NewMicrofrontendService(informer cache.SharedIndexInformer) *microfrontendService {
	return newMicrofrontendService(informer)
}

func NewMicrofrontendResolver(microfrontendSvc microfrontendSvc) *microfrontendResolver {
	return newMicrofrontendResolver(microfrontendSvc)
}

func (r *microfrontendResolver) SetMicrofrontendConverter(converter gqlMicrofrontendConverter) {
	r.microfrontendConverter = converter
}
