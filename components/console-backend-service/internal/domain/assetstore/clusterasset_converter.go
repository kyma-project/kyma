package assetstore

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlClusterAssetConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlClusterAssetConverter -case=underscore -output disabled -outpkg disabled
type gqlClusterAssetConverter interface {
	ToGQL(in *v1alpha2.ClusterAsset) (*gqlschema.ClusterAsset, error)
	ToGQLs(in []*v1alpha2.ClusterAsset) ([]gqlschema.ClusterAsset, error)
}

type clusterAssetConverter struct {
	extractor extractor.AssetStatusExtractor
}

func (c *clusterAssetConverter) ToGQL(item *v1alpha2.ClusterAsset) (*gqlschema.ClusterAsset, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonAssetStatus)

	clusterAsset := gqlschema.ClusterAsset{
		Name:   item.Name,
		Type:   item.Labels[CmsTypeLabel],
		Status: status,
	}

	return &clusterAsset, nil
}

func (c *clusterAssetConverter) ToGQLs(in []*v1alpha2.ClusterAsset) ([]gqlschema.ClusterAsset, error) {
	var result []gqlschema.ClusterAsset
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}
