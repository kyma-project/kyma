package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type asset struct {
	dynamicCli dynamic.Interface
	res        *resource.Resource
	BucketName string
	Namespace string
}

func newAsset(dynamicCli dynamic.Interface, namespace string) *asset {
	return &asset{
		res: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "assets",
		}, namespace),
		dynamicCli: dynamicCli,
		Namespace:namespace,
	}
}

func (a *asset) Create(assets []assetData) error {
	for _, asset := range assets {
		asset := &v1alpha2.Asset{
			TypeMeta: metav1.TypeMeta{
				Kind: "Asset",
				APIVersion: v1alpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      asset.Name,
				Namespace: a.Namespace,
			},
			Spec:v1alpha2.AssetSpec{
				CommonAssetSpec: v1alpha2.CommonAssetSpec{
					BucketRef: v1alpha2.AssetBucketRef{
						Name: a.BucketName,
					},
					Source:v1alpha2.AssetSource{
						Url: asset.URL,
						Mode:asset.Mode,
					},
				},
			},
		}

		err := a.res.Create(asset)
		if err != nil {
			return errors.Wrapf(err, "while creating Asset %s in namespace %s", asset.Name, a.Namespace)
		}
	}

	return nil
}

func (a *asset) Delete(assets []assetData) error {
	for _, asset := range assets {
		err := a.res.Delete(asset.Name)
		if err != nil {
			return errors.Wrapf(err, "while deleting Asset %s in namespace %s", asset.Name, a.Namespace)
		}
	}

	return nil
}

