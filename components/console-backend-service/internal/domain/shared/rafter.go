package shared

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/spec"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

//go:generate mockery -name=RafterRetriever -output=automock -outpkg=automock -case=underscore
type RafterRetriever interface {
	ClusterAssetGroup() ClusterAssetGroupGetter
	AssetGroup() AssetGroupGetter
	ClusterAssetGroupConverter() GqlClusterAssetGroupConverter
	AssetGroupConverter() GqlAssetGroupConverter
	ClusterAsset() ClusterAssetGetter
	Asset() AssetGetter
	ClusterAssetConverter() GqlClusterAssetConverter
	AssetConverter() GqlAssetConverter
	Specification() SpecificationGetter
}

//go:generate mockery -name=ClusterAssetGroupGetter -output=automock -outpkg=automock -case=underscore
type ClusterAssetGroupGetter interface {
	Find(name string) (*v1beta1.ClusterAssetGroup, error)
}

//go:generate mockery -name=AssetGroupGetter -output=automock -outpkg=automock -case=underscore
type AssetGroupGetter interface {
	Find(namespace, name string) (*v1beta1.AssetGroup, error)
}

//go:generate mockery -name=GqlClusterAssetGroupConverter -output=automock -outpkg=automock -case=underscore
type GqlClusterAssetGroupConverter interface {
	ToGQL(item *v1beta1.ClusterAssetGroup) (*gqlschema.ClusterAssetGroup, error)
	ToGQLs(in []*v1beta1.ClusterAssetGroup) ([]gqlschema.ClusterAssetGroup, error)
}

//go:generate mockery -name=GqlAssetGroupConverter -output=automock -outpkg=automock -case=underscore
type GqlAssetGroupConverter interface {
	ToGQL(item *v1beta1.AssetGroup) (*gqlschema.AssetGroup, error)
	ToGQLs(in []*v1beta1.AssetGroup) ([]gqlschema.AssetGroup, error)
}

//go:generate mockery -name=ClusterAssetGetter -output=automock -outpkg=automock -case=underscore
type ClusterAssetGetter interface {
	ListForClusterAssetGroupByType(assetGroupName string, types []string) ([]*v1beta1.ClusterAsset, error)
}

//go:generate mockery -name=AssetGetter -output=automock -outpkg=automock -case=underscore
type AssetGetter interface {
	ListForAssetGroupByType(namespace, assetGroupName string, types []string) ([]*v1beta1.Asset, error)
}

//go:generate mockery -name=GqlClusterAssetConverter -output=automock -outpkg=automock -case=underscore
type GqlClusterAssetConverter interface {
	ToGQL(item *v1beta1.ClusterAsset) (*gqlschema.ClusterAsset, error)
	ToGQLs(in []*v1beta1.ClusterAsset) ([]gqlschema.ClusterAsset, error)
}

//go:generate mockery -name=GqlAssetConverter -output=automock -outpkg=automock -case=underscore
type GqlAssetConverter interface {
	ToGQL(item *v1beta1.Asset) (*gqlschema.Asset, error)
	ToGQLs(in []*v1beta1.Asset) ([]gqlschema.Asset, error)
}

//go:generate mockery -name=SpecificationGetter -output=automock -outpkg=automock -case=underscore
type SpecificationGetter interface {
	AsyncAPI(baseURL, name string) (*spec.AsyncAPISpec, error)
}
