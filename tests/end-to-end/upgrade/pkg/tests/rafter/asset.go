package rafter

import (
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/dynamicresource"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type asset struct {
	resCli    *dynamicresource.DynamicResource
	name      string
	namespace string
	data      assetData
}

func newAsset(dynamicCli dynamic.Interface, namespace string, data assetData) *asset {
	return &asset{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assets",
		}),
		name:      data.name,
		namespace: namespace,
		data:      data,
	}
}

func (a *asset) create() error {
	asset := &v1beta1.Asset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Asset",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.data.name,
			Namespace: a.namespace,
		},
		Spec: v1beta1.AssetSpec{
			CommonAssetSpec: v1beta1.CommonAssetSpec{
				BucketRef: v1beta1.AssetBucketRef{
					Name: bucketName,
				},
				Source: v1beta1.AssetSource{
					URL:  a.data.url,
					Mode: a.data.mode,
				},
			},
		},
	}

	err := a.resCli.Create(asset)
	if err != nil {
		return errors.Wrapf(err, "while creating Asset %s in namespace %s", asset.Name, a.namespace)
	}

	return nil
}

func (a *asset) get() (*v1beta1.Asset, error) {
	var res v1beta1.Asset
	err := a.resCli.Get(a.namespace, a.name, &res)
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

		if res.Status.Phase != v1beta1.AssetReady {
			return false, nil
		}

		return true, nil
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Asset %s in namespace %s", a.name, a.namespace)
	}

	return nil
}
