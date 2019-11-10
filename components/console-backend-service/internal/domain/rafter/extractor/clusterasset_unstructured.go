package extractor

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type ClusterAssetUnstructuredExtractor struct{}

func (e *ClusterAssetUnstructuredExtractor) Do(obj interface{}) (*v1beta1.ClusterAsset, error) {
	u, err := resource.ToUnstructured(obj)
	if err != nil || u == nil {
		return nil, err
	}

	clusterAsset := &v1beta1.ClusterAsset{}
	err = resource.FromUnstructured(u, clusterAsset)
	return clusterAsset, err
}
