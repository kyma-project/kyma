package rafter

import "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

type retriever struct {
	ClusterAssetGroupGetter       shared.ClusterAssetGroupGetter
	AssetGroupGetter              shared.AssetGroupGetter
	GqlClusterAssetGroupConverter shared.GqlClusterAssetGroupConverter
	GqlAssetGroupConverter        shared.GqlAssetGroupConverter
	ClusterAssetGetter            shared.RafterClusterAssetGetter
	SpecificationSvc              shared.RafterSpecificationGetter
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

func (r *retriever) ClusterAsset() shared.RafterClusterAssetGetter {
	return r.ClusterAssetGetter
}

func (r *retriever) Specification() shared.RafterSpecificationGetter {
	return r.SpecificationSvc
}
