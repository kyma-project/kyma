package rafter

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlClusterAssetGroupConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlClusterAssetGroupConverter -case=underscore -output disabled -outpkg disabled
type gqlClusterAssetGroupConverter interface {
	ToGQL(in *v1beta1.ClusterAssetGroup) (*gqlschema.ClusterAssetGroup, error)
	ToGQLs(in []*v1beta1.ClusterAssetGroup) ([]gqlschema.ClusterAssetGroup, error)
}

type clusterAssetGroupConverter struct {
	extractor extractor.AssetGroupCommonExtractor
}

func newClusterAssetGroupConverter() *clusterAssetGroupConverter {
	return &clusterAssetGroupConverter{
		extractor: extractor.AssetGroupCommonExtractor{},
	}
}

func (c *clusterAssetGroupConverter) ToGQL(item *v1beta1.ClusterAssetGroup) (*gqlschema.ClusterAssetGroup, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonAssetGroupStatus)
	clusterAssetGroup := gqlschema.ClusterAssetGroup{
		Name:        item.Name,
		Description: item.Spec.Description,
		DisplayName: item.Spec.DisplayName,
		GroupName:   item.Labels[GroupNameLabel],
		Status:      status,
	}

	return &clusterAssetGroup, nil
}

func (c *clusterAssetGroupConverter) ToGQLs(in []*v1beta1.ClusterAssetGroup) ([]gqlschema.ClusterAssetGroup, error) {
	var result []gqlschema.ClusterAssetGroup
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
