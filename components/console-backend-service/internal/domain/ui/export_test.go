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

func NewMicrofrontendResolver(microfrontendLister microfrontendLister) *microfrontendResolver {
	return newMicrofrontendResolver(microfrontendLister)
}

func (r *microfrontendResolver) SetMicrofrontendConverter(converter gqlMicrofrontendConverter) {
	r.microfrontendConverter = converter
}

func NewClusterMicrofrontendService(informer cache.SharedIndexInformer) *clusterMicrofrontendService {
	return newClusterMicrofrontendService(informer)
}

func NewClusterMicrofrontendResolver(clusterMicrofrontendLister clusterMicrofrontendLister) *clusterMicrofrontendResolver {
	return newClusterMicrofrontendResolver(clusterMicrofrontendLister)
}

func (r *clusterMicrofrontendResolver) SetClusterMicrofrontendConverter(converter gqlClusterMicrofrontendConverter) {
	r.clusterMicrofrontendConverter = converter
}
