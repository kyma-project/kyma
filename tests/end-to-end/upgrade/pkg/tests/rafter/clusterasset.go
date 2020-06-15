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

type clusterAsset struct {
	resCli *dynamicresource.DynamicResource
	name   string
	data   assetData
}

func newClusterAsset(dynamicCli dynamic.Interface, data assetData) *clusterAsset {
	return &clusterAsset{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "clusterassets",
		}),
		name: fixSimpleAssetData().name,
		data: data,
	}
}

func (a *clusterAsset) create() error {
	asset := &v1beta1.ClusterAsset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAsset",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: a.data.name,
		},
		Spec: v1beta1.ClusterAssetSpec{
			CommonAssetSpec: v1beta1.CommonAssetSpec{
				BucketRef: v1beta1.AssetBucketRef{
					Name: clusterBucketName,
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
		return errors.Wrapf(err, "while creating ClusterAsset %s", asset.Name)
	}

	return nil
}

func (a *clusterAsset) get() (*v1beta1.ClusterAsset, error) {
	var ca v1beta1.ClusterAsset
	err := a.resCli.Get("", a.name, &ca)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterAsset %s", a.name)
	}

	return &ca, nil
}

func (a *clusterAsset) waitForStatusReady(stop <-chan struct{}) error {
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
		return errors.Wrapf(err, "while waiting for ready ClusterAsset %s", a.name)
	}

	return nil
}
