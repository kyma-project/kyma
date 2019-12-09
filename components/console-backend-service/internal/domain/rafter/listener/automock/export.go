package automock

func NewGqlAssetConverter() *gqlAssetConverter {
	return new(gqlAssetConverter)
}

func NewGqlAssetGroupConverter() *gqlAssetGroupConverter {
	return new(gqlAssetGroupConverter)
}

func NewGqlClusterAssetConverter() *gqlClusterAssetConverter {
	return new(gqlClusterAssetConverter)
}

func NewGqlClusterAssetGroupConverter() *gqlClusterAssetGroupConverter {
	return new(gqlClusterAssetGroupConverter)
}
