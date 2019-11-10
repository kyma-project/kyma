package rafter

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

func NewClusterAssetGroupResolver(clusterAssetGroupService clusterAssetGroupSvc, clusterAssetGroupConverter gqlClusterAssetGroupConverter, clusterAssetService clusterAssetSvc, clusterAssetConverter gqlClusterAssetConverter) *clusterAssetGroupResolver {
	return newClusterAssetGroupResolver(clusterAssetGroupService, clusterAssetGroupConverter, clusterAssetService, clusterAssetConverter)
}

func NewClusterAssetGroupService(serviceFactory *resource.ServiceFactory) (*clusterAssetGroupService, error) {
	return newClusterAssetGroupService(serviceFactory)
}

func NewAssetGroupResolver(assetGroupService assetGroupSvc, assetGroupConverter gqlAssetGroupConverter, assetService assetSvc, assetConverter gqlAssetConverter) *assetGroupResolver {
	return newAssetGroupResolver(assetGroupService, assetGroupConverter, assetService, assetConverter)
}

func NewAssetGroupService(serviceFactory *resource.ServiceFactory) (*assetGroupService, error) {
	return newAssetGroupService(serviceFactory)
}

func NewFileService() *fileService {
	return newFileService()
}

func NewFileConverter() *fileConverter {
	return newFileConverter()
}
