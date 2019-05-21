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

func NewMicroFrontendService(informer cache.SharedIndexInformer) *microFrontendService {
	return newMicroFrontendService(informer)
}

func NewMicroFrontendResolver(microFrontendLister microFrontendLister) *microFrontendResolver {
	return newMicroFrontendResolver(microFrontendLister)
}

func (r *microFrontendResolver) SetMicroFrontendConverter(converter gqlMicroFrontendConverter) {
	r.microFrontendConverter = converter
}

func NewClusterMicroFrontendService(informer cache.SharedIndexInformer) *clusterMicroFrontendService {
	return newClusterMicroFrontendService(informer)
}

func NewClusterMicroFrontendResolver(clusterMicroFrontendLister clusterMicroFrontendLister) *clusterMicroFrontendResolver {
	return newClusterMicroFrontendResolver(clusterMicroFrontendLister)
}

func (r *clusterMicroFrontendResolver) SetClusterMicroFrontendConverter(converter gqlClusterMicroFrontendConverter) {
	r.clusterMicroFrontendConverter = converter
}
