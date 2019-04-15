package shared

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=AssetStoreRetriever -output=automock -outpkg=automock -case=underscore
type AssetStoreRetriever interface {
	ClusterAsset() ClusterAssetGetter
	Asset() AssetGetter
	ClusterAssetConverter() GqlClusterAssetConverter
	AssetConverter() GqlAssetConverter
}

//go:generate mockery -name=ClusterAssetGetter -output=automock -outpkg=automock -case=underscore
type ClusterAssetGetter interface {
	ListForDocsTopicByType(docsTopicName string, types []string) ([]*v1alpha2.ClusterAsset, error)
}

//go:generate mockery -name=AssetGetter -output=automock -outpkg=automock -case=underscore
type AssetGetter interface {
	ListForDocsTopicByType(namespace, docsTopicName string, types []string) ([]*v1alpha2.Asset, error)
}

//go:generate mockery -name=GqlClusterAssetConverter -output=automock -outpkg=automock -case=underscore
type GqlClusterAssetConverter interface {
	ToGQL(item *v1alpha2.ClusterAsset) (*gqlschema.ClusterAsset, error)
	ToGQLs(in []*v1alpha2.ClusterAsset) ([]gqlschema.ClusterAsset, error)
}

//go:generate mockery -name=GqlAssetConverter -output=automock -outpkg=automock -case=underscore
type GqlAssetConverter interface {
	ToGQL(item *v1alpha2.Asset) (*gqlschema.Asset, error)
	ToGQLs(in []*v1alpha2.Asset) ([]gqlschema.Asset, error)
}
