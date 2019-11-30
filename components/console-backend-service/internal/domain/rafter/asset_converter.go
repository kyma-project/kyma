package rafter

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlAssetConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlAssetConverter -case=underscore -output disabled -outpkg disabled
type gqlAssetConverter interface {
	ToGQL(in *v1beta1.Asset) (*gqlschema.Asset, error)
	ToGQLs(in []*v1beta1.Asset) ([]gqlschema.Asset, error)
}

type assetConverter struct {
	extractor extractor.AssetCommonExtractor
}

func newAssetConverter() *assetConverter {
	return &assetConverter{
		extractor: extractor.AssetCommonExtractor{},
	}
}

func (c *assetConverter) ToGQL(item *v1beta1.Asset) (*gqlschema.Asset, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonAssetStatus)
	parameters, err := c.extractor.Parameters(item.Spec.CommonAssetSpec.Parameters)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting parameters from %s [name: %s][namespace: %s]", pretty.AssetType, item.Name, item.Namespace)
	}

	asset := gqlschema.Asset{
		Name:       item.Name,
		Namespace:  item.Namespace,
		Type:       item.Labels[TypeLabel],
		Status:     status,
		Metadata:   parameters,
		Parameters: parameters,
	}

	return &asset, nil
}

func (c *assetConverter) ToGQLs(in []*v1beta1.Asset) ([]gqlschema.Asset, error) {
	var result []gqlschema.Asset
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
