package automock

func NewClusterAssetService() *clusterAssetSvc {
	return new(clusterAssetSvc)
}

func NewGQLClusterAssetConverter() *gqlClusterAssetConverter {
	return new(gqlClusterAssetConverter)
}

func NewAssetService() *assetSvc {
	return new(assetSvc)
}

func NewGQLAssetConverter() *gqlAssetConverter {
	return new(gqlAssetConverter)
}

func NewClusterAssetGroupService() *clusterAssetGroupSvc {
	return new(clusterAssetGroupSvc)
}

func NewGQLClusterAssetGroupConverter() *gqlClusterAssetGroupConverter {
	return new(gqlClusterAssetGroupConverter)
}

func NewAssetGroupService() *assetGroupSvc {
	return new(assetGroupSvc)
}

func NewGQLAssetGroupConverter() *gqlAssetGroupConverter {
	return new(gqlAssetGroupConverter)
}

func NewFileService() *fileSvc {
	return new(fileSvc)
}

func NewGQLFileConverter() *gqlFileConverter {
	return new(gqlFileConverter)
}