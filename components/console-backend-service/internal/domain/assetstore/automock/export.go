package automock

// ClusterAsset

func NewClusterAssetService() *clusterAssetSvc {
	return new(clusterAssetSvc)
}

func NewGQLClusterAssetConverter() *gqlClusterAssetConverter {
	return new(gqlClusterAssetConverter)
}

// Asset

func NewAssetService() *assetSvc {
	return new(assetSvc)
}

func NewGQLAssetConverter() *gqlAssetConverter {
	return new(gqlAssetConverter)
}

// File

func NewFileService() *fileSvc {
	return new(fileSvc)
}

func NewGQLFileConverter() *gqlFileConverter {
	return new(gqlFileConverter)
}