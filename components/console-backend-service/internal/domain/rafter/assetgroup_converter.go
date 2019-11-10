package rafter

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=gqlAssetGroupConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlAssetGroupConverter -case=underscore -output disabled -outpkg disabled
type gqlAssetGroupConverter interface {
	ToGQL(in *v1beta1.AssetGroup) (*gqlschema.AssetGroup, error)
	ToGQLs(in []*v1beta1.AssetGroup) ([]gqlschema.AssetGroup, error)
}

type assetGroupConverter struct {
	extractor extractor.AssetGroupCommonExtractor
}

func newAssetGroupConverter() *assetGroupConverter {
	return &assetGroupConverter{
		extractor: extractor.AssetGroupCommonExtractor{},
	}
}

func (c *assetGroupConverter) ToGQL(item *v1beta1.AssetGroup) (*gqlschema.AssetGroup, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonAssetGroupStatus)
	assetGroup := gqlschema.AssetGroup{
		Name:        item.Name,
		Namespace:   item.Namespace,
		Description: item.Spec.Description,
		DisplayName: item.Spec.DisplayName,
		GroupName:   item.Labels[GroupNameLabel],
		Status:      status,
	}

	return &assetGroup, nil
}

func (c *assetGroupConverter) ToGQLs(in []*v1beta1.AssetGroup) ([]gqlschema.AssetGroup, error) {
	var result []gqlschema.AssetGroup
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
