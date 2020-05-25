package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource"
)

func NewBackendModuleResolver(sf *resource.ServiceFactory) *backendModuleResolver {
	return newBackendModuleResolver(sf)
}

func (r *backendModuleResolver) SetInstanceConverter(converter gqlBackendModuleConverter) {
	r.converter = converter
}

//func NewMicroFrontendResolver(microFrontendLister microFrontendLister) *microFrontendResolver {
//	return newMicroFrontendResolver(microFrontendLister)
//}
//
//func (r *microFrontendResolver) SetMicroFrontendConverter(converter gqlMicroFrontendConverter) {
//	r.microFrontendConverter = converter
//}
//
//func NewClusterMicroFrontendResolver(clusterMicroFrontendLister clusterMicroFrontendLister) *clusterMicroFrontendResolver {
//	return newClusterMicroFrontendResolver(clusterMicroFrontendLister)
//}
//
//func (r *clusterMicroFrontendResolver) SetClusterMicroFrontendConverter(converter gqlClusterMicroFrontendConverter) {
//	r.clusterMicroFrontendConverter = converter
//}
