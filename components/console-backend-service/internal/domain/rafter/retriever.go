package rafter

import "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

type retriever struct {
	ClusterAssetGroupGetter       shared.ClusterAssetGroupGetter
	AssetGroupGetter              shared.AssetGroupGetter
	GqlClusterAssetGroupConverter shared.GqlClusterAssetGroupConverter
	GqlAssetGroupConverter        shared.GqlAssetGroupConverter
	ClusterAssetGetter            shared.ClusterAssetGetter
	AssetGetter                   shared.AssetGetter
	GqlClusterAssetConverter      shared.GqlClusterAssetConverter
	GqlAssetConverter             shared.GqlAssetConverter
	SpecificationSvc              shared.SpecificationGetter
}

func (r *retriever) ClusterAssetGroup() shared.ClusterAssetGroupGetter {
	return r.ClusterAssetGroupGetter
}

func (r *retriever) AssetGroup() shared.AssetGroupGetter {
	return r.AssetGroupGetter
}

func (r *retriever) ClusterAssetGroupConverter() shared.GqlClusterAssetGroupConverter {
	return r.GqlClusterAssetGroupConverter
}

func (r *retriever) AssetGroupConverter() shared.GqlAssetGroupConverter {
	return r.GqlAssetGroupConverter
}

func (r *retriever) ClusterAsset() shared.ClusterAssetGetter {
	return r.ClusterAssetGetter
}

func (r *retriever) Asset() shared.AssetGetter {
	return r.AssetGetter
}

func (r *retriever) ClusterAssetConverter() shared.GqlClusterAssetConverter {
	return r.GqlClusterAssetConverter
}

func (r *retriever) AssetConverter() shared.GqlAssetConverter {
	return r.GqlAssetConverter
}

func (r *retriever) Specification() shared.SpecificationGetter {
	return r.SpecificationSvc
}
