package rafter

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
)

//go:generate mockery -name=gqlClusterAssetConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlClusterAssetConverter -case=underscore -output disabled -outpkg disabled
type gqlClusterAssetConverter interface {
	ToGQL(in *v1beta1.ClusterAsset) (*gqlschema.ClusterAsset, error)
	ToGQLs(in []*v1beta1.ClusterAsset) ([]gqlschema.ClusterAsset, error)
}

type clusterAssetConverter struct {
	extractor extractor.AssetCommonExtractor
}

func newClusterAssetConverter() *clusterAssetConverter {
	return &clusterAssetConverter{
		extractor: extractor.AssetCommonExtractor{},
	}
}

func (c *clusterAssetConverter) ToGQL(item *v1beta1.ClusterAsset) (*gqlschema.ClusterAsset, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonAssetStatus)
	parameters, err := c.extractor.Parameters(item.Spec.CommonAssetSpec.Parameters)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting parameters from %s [name: %s]", pretty.ClusterAssetType, item.Name)
	}

	clusterAsset := gqlschema.ClusterAsset{
		Name:       item.Name,
		Type:       item.Labels[TypeLabel],
		Status:     status,
		Parameters: parameters,
	}

	return &clusterAsset, nil
}

func (c *clusterAssetConverter) ToGQLs(in []*v1beta1.ClusterAsset) ([]gqlschema.ClusterAsset, error) {
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
