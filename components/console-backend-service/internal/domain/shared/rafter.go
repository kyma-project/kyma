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
	ClusterAsset() RafterClusterAssetGetter
	Specification() RafterSpecificationGetter
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

//go:generate mockery -name=RafterClusterAssetGetter -output=automock -outpkg=automock -case=underscore
type RafterClusterAssetGetter interface {
	ListForClusterAssetGroupByType(assetGroupName string, types []string) ([]*v1beta1.ClusterAsset, error)
}

//go:generate mockery -name=RafterSpecificationGetter -output=automock -outpkg=automock -case=underscore
type RafterSpecificationGetter interface {
	AsyncAPI(baseURL, name string) (*spec.AsyncAPISpec, error)
}
