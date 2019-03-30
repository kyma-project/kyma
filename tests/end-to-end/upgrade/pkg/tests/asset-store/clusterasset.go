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

type clusterAsset struct {
	resCli            *resource.Resource
	name 			  string
}

func newClusterAssetClient(dynamicCli dynamic.Interface) *clusterAsset {
	return &clusterAsset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "clusterassets",
		}, ""),
	}
}

func (a *clusterAsset) create() error {
	assetData := fixSimpleAssetData()

	asset := &v1alpha2.ClusterAsset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAsset",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: assetData.name,
		},
		Spec: v1alpha2.ClusterAssetSpec{
			CommonAssetSpec: v1alpha2.CommonAssetSpec{
				BucketRef: v1alpha2.AssetBucketRef{
					Name: ClusterBucketName,
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
		return errors.Wrapf(err, "while creating ClusterAsset %s", asset.Name)
	}

	a.name = asset.Name
	return nil
}

func (a *clusterAsset) get() (*v1alpha2.ClusterAsset, error) {
	u, err := a.resCli.Get(a.name)
	if err != nil {
		return nil, err
	}

	var ca v1alpha2.ClusterAsset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ca)
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

		if res.Status.Phase != v1alpha2.AssetReady {
			return false, nil
		}

		return true, nil
	}, WaitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready ClusterAsset resource")
	}

	return nil
}