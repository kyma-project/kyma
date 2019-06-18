package assetstore

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

//go:generate mockery -name=gqlAssetConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlAssetConverter -case=underscore -output disabled -outpkg disabled
type gqlAssetConverter interface {
	ToGQL(in *v1alpha2.Asset) (*gqlschema.Asset, error)
	ToGQLs(in []*v1alpha2.Asset) ([]gqlschema.Asset, error)
}

type assetConverter struct {
	extractor extractor.AssetStatusExtractor
}

func (c *assetConverter) ToGQL(item *v1alpha2.Asset) (*gqlschema.Asset, error) {
	if item == nil {
		return nil, nil
	}

	status := c.extractor.Status(item.Status.CommonAssetStatus)
	metadata, err := c.extractMetadata(item.Spec.Metadata)
	if err != nil {
		return nil, errors.Wrapf(err, "while extracting metadata from Asset [name: %s][namespace: %s]", item.Name, item.Namespace)
	}

	asset := gqlschema.Asset{
		Name:      item.Name,
		Namespace: item.Namespace,
		Type:      item.Labels[CmsTypeLabel],
		Status:    status,
		Metadata:  metadata,
	}

	return &asset, nil
}

func (c *assetConverter) ToGQLs(in []*v1alpha2.Asset) ([]gqlschema.Asset, error) {
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

func (*assetConverter) extractMetadata(ext *runtime.RawExtension) (map[string]interface{}, error) {
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
