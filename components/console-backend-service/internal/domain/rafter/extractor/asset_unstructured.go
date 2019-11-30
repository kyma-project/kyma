package extractor

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type AssetUnstructuredExtractor struct{}

func (e *AssetUnstructuredExtractor) Do(obj interface{}) (*v1beta1.Asset, error) {
	u, err := resource.ToUnstructured(obj)
	if err != nil || u == nil {
		return nil, err
	}

	asset := &v1beta1.Asset{}
	err = resource.FromUnstructured(u, asset)
	return asset, err
}
