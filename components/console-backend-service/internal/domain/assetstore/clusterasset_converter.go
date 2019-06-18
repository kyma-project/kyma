package assetstore

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	metadata, err := c.extractMetadata(item.Spec.Metadata)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting metadata from ClusterAsset [name: %s]", item.Name)
	}

	clusterAsset := gqlschema.ClusterAsset{
		Name:     item.Name,
		Type:     item.Labels[CmsTypeLabel],
		Status:   status,
		Metadata: metadata,
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

func (*clusterAssetConverter) extractMetadata(ext *runtime.RawExtension) (map[string]interface{}, error) {
	if ext == nil {
		return nil, nil
	}
	result := make(map[string]interface{})
	err := json.Unmarshal(ext.Raw, &result)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling metadata")
	}

	return result, nil
}
