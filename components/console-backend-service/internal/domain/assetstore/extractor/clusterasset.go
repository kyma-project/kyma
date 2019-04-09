package extractor

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/pretty"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterAssetUnstructuredExtractor struct{}

func (ext *ClusterAssetUnstructuredExtractor) Do(obj interface{}) (*v1alpha2.ClusterAsset, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.ClusterAsset, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	var clusterAsset v1alpha2.ClusterAsset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &clusterAsset)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.ClusterAsset, u)
	}

	return &clusterAsset, nil
}
