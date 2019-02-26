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
	assets []assetDetails
	Namespace string
}

func newClusterAsset(dynamicCli dynamic.Interface, assets []assetDetails, namespace string) *clusterAsset {
	return &clusterAsset{
		res: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "assets",
		}, namespace),
		dynamicCli: dynamicCli,
		Namespace:namespace,
	}
}

func (a *clusterAsset) Create() error {
	for _, asset := range a.assets {
		asset := &v1alpha2.ClusterAsset{
			TypeMeta: metav1.TypeMeta{
				Kind: "ClusterAsset",
				APIVersion: v1alpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      asset.Name,
				Namespace: a.Namespace,
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
			return errors.Wrapf(err, "while creating asset %s in namespace %s", asset.Name, a.Namespace)
		}
	}

	return nil
}

func (a *clusterAsset) Delete() error {
	for _, asset := range a.assets {
		err := a.res.Delete(asset.Name)
		if err != nil {
			return errors.Wrapf(err, "while deleting asset %s in namespace %s", asset.Name, a.Namespace)
		}
	}

	return nil
}
