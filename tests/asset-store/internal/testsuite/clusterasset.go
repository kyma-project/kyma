package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type clusterAsset struct {
	dynamicCli dynamic.Interface
	res        *resource.Resource
}

func newClusterAsset(dynamicCli dynamic.Interface) *clusterAsset {
	return &clusterAsset{
		res: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "clusterassets",
		}, ""),
		dynamicCli: dynamicCli,
	}
}

func (a *clusterAsset) Create(assets []assetDetails) error {
	for _, asset := range assets {
		asset := &v1alpha2.ClusterAsset{
			TypeMeta: metav1.TypeMeta{
				Kind: "ClusterAsset",
				APIVersion: v1alpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      asset.Name,
			},
			Spec:v1alpha2.ClusterAssetSpec{
				CommonAssetSpec: v1alpha2.CommonAssetSpec{
					Source:v1alpha2.AssetSource{
						Url: asset.URL,
						Mode:asset.Mode,
					},
				},
			},
		}

		err := a.res.Create(asset)
		if err != nil {
			return errors.Wrapf(err, "while creating ClusterAsset %s", asset.Name)
		}
	}

	return nil
}

func (a *clusterAsset) Delete(assets []assetDetails) error {
	for _, asset := range assets {
		err := a.res.Delete(asset.Name)
		if err != nil {
			return errors.Wrapf(err, "while deleting ClusterAsset %s", asset.Name)
		}
	}

	return nil
}
