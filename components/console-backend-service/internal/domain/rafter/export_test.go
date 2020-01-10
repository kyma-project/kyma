package rafter

import (
	"net/http"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

func NewClusterAssetResolver(clusterAssetService clusterAssetSvc, clusterAssetConverter gqlClusterAssetConverter, fileService fileSvc, fileConverter gqlFileConverter) *clusterAssetResolver {
	return newClusterAssetResolver(clusterAssetService, clusterAssetConverter, fileService, fileConverter)
}

type ClusterAssetService = clusterAssetService

func NewClusterAssetService(serviceFactory *resource.ServiceFactory) (*clusterAssetService, error) {
	return newClusterAssetService(serviceFactory)
}

func NewClusterAssetConverter() *clusterAssetConverter {
	return newClusterAssetConverter()
}

func NewAssetResolver(assetService assetSvc, assetConverter gqlAssetConverter, fileService fileSvc, fileConverter gqlFileConverter) *assetResolver {
	return newAssetResolver(assetService, assetConverter, fileService, fileConverter)
}

type AssetService = assetService

func NewAssetService(serviceFactory *resource.ServiceFactory) (*assetService, error) {
	return newAssetService(serviceFactory)
}

func NewAssetConverter() *assetConverter {
	return newAssetConverter()
}

func NewClusterAssetGroupResolver(clusterAssetGroupService clusterAssetGroupSvc, clusterAssetGroupConverter gqlClusterAssetGroupConverter, clusterAssetService clusterAssetSvc, clusterAssetConverter gqlClusterAssetConverter) *clusterAssetGroupResolver {
	return newClusterAssetGroupResolver(clusterAssetGroupService, clusterAssetGroupConverter, clusterAssetService, clusterAssetConverter)
}

type ClusterAssetGroupService = clusterAssetGroupService

func NewClusterAssetGroupService(serviceFactory *resource.ServiceFactory) (*clusterAssetGroupService, error) {
	return newClusterAssetGroupService(serviceFactory)
}

func NewClusterAssetGroupConverter() *clusterAssetGroupConverter {
	return newClusterAssetGroupConverter()
}

func NewAssetGroupResolver(assetGroupService assetGroupSvc, assetGroupConverter gqlAssetGroupConverter, assetService assetSvc, assetConverter gqlAssetConverter) *assetGroupResolver {
	return newAssetGroupResolver(assetGroupService, assetGroupConverter, assetService, assetConverter)
}

type AssetGroupService = assetGroupService

func NewAssetGroupService(serviceFactory *resource.ServiceFactory) (*assetGroupService, error) {
	return newAssetGroupService(serviceFactory)
}

func NewAssetGroupConverter() *assetGroupConverter {
	return newAssetGroupConverter()
}

func NewFileService() *fileService {
	return newFileService()
}

func NewFileConverter() *fileConverter {
	return newFileConverter()
}

type SpecificationService = specificationService

func NewSpecificationService(cfg Config, endpoint string, client *http.Client) *SpecificationService {
	return &SpecificationService{
		cfg:      cfg,
		endpoint: endpoint,
		client:   client,
	}
}

func (s *SpecificationService) ReadData(baseURL, name string) ([]byte, error) {
	return s.readData(baseURL, name)
}

func (s *SpecificationService) PreparePath(baseURL, name string) string {
	return s.preparePath(baseURL, name)
}

func (s *SpecificationService) Fetch(url string) ([]byte, error) {
	return s.fetch(url)
}
