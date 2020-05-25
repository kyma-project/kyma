package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource"
)

func NewBackendModuleResolver(sf *resource.ServiceFactory) *backendModuleResolver {
	return newBackendModuleResolver(sf)
}

func NewMicroFrontendResolver(sf *resource.ServiceFactory) *microFrontendResolver {
	return newMicroFrontendResolver(sf)
}

//
//func NewClusterMicroFrontendResolver(clusterMicroFrontendLister clusterMicroFrontendLister) *clusterMicroFrontendResolver {
//	return newClusterMicroFrontendResolver(clusterMicroFrontendLister)
//}
//
//func (r *clusterMicroFrontendResolver) SetClusterMicroFrontendConverter(converter gqlClusterMicroFrontendConverter) {
//	r.clusterMicroFrontendConverter = converter
//}
