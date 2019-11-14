package extractor

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type AssetGroupUnstructuredExtractor struct{}

func (e *AssetGroupUnstructuredExtractor) Do(obj interface{}) (*v1beta1.AssetGroup, error) {
	u, err := resource.ToUnstructured(obj)
	if err != nil || u == nil {
		return nil, err
	}

	assetGroup := &v1beta1.AssetGroup{}
	err = resource.FromUnstructured(u, assetGroup)
	return assetGroup, err
}
