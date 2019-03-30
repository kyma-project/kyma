package asset_store

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/resource"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
)

type asset struct {
	resCli      *resource.Resource
	name        string
	namespace   string
}

func newAssetClient(dynamicCli dynamic.Interface, namespace string) *asset {
	return &asset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "assets",
		}, namespace),
		namespace:   namespace,
	}
}

func (a *asset) create() error {
	assetData := fixSimpleAssetData()

	asset := &v1alpha2.Asset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Asset",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      assetData.name,
			Namespace: a.namespace,
		},
		Spec: v1alpha2.AssetSpec{
			CommonAssetSpec: v1alpha2.CommonAssetSpec{
				BucketRef: v1alpha2.AssetBucketRef{
					Name: BucketName,
				},
				Source: v1alpha2.AssetSource{
					Url:  assetData.url,
					Mode: assetData.mode,
				},
			},
		},
	}

	err := a.resCli.Create(asset)
	if err != nil {
		return errors.Wrapf(err, "while creating Asset %s in namespace %s", asset.Name, a.namespace)
	}

	a.name = asset.Name
	return nil
}

func (a *asset) get() (*v1alpha2.Asset, error) {
	u, err := a.resCli.Get(a.name)
	if err != nil {
		return nil, err
	}

	var res v1alpha2.Asset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Asset %s in namespace %s", a.name, a.namespace)
	}

	return &res, nil
}

func (a *asset) waitForStatusReady(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := a.get()
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1alpha2.AssetReady {
			return false, nil
		}

		return true, nil
	}, WaitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Asset resource")
	}

	return nil
}