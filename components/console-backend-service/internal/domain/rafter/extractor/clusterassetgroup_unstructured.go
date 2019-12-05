package extractor

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type ClusterAssetGroupUnstructuredExtractor struct{}

func (e *ClusterAssetGroupUnstructuredExtractor) Do(obj interface{}) (*v1beta1.ClusterAssetGroup, error) {
	u, err := resource.ToUnstructured(obj)
	if err != nil || u == nil {
		return nil, err
	}

	clusterAssetGroup := &v1beta1.ClusterAssetGroup{}
	err = resource.FromUnstructured(u, clusterAssetGroup)
	return clusterAssetGroup, err
}
