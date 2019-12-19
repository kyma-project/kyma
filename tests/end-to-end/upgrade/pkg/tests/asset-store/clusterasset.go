package assetstore

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/dynamicresource"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type clusterAsset struct {
	resCli *dynamicresource.DynamicResource
	name   string
}

func newClusterAsset(dynamicCli dynamic.Interface) *clusterAsset {
	return &clusterAsset{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.GroupVersion.Version,
			Group:    v1alpha2.GroupVersion.Group,
			Resource: "clusterassets",
		}, ""),
		name: fixSimpleAssetData().name,
	}
}

func (a *clusterAsset) create() error {
	assetData := fixSimpleAssetData()

	asset := &v1alpha2.ClusterAsset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAsset",
			APIVersion: v1alpha2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: assetData.name,
		},
		Spec: v1alpha2.ClusterAssetSpec{
			CommonAssetSpec: v1alpha2.CommonAssetSpec{
				BucketRef: v1alpha2.AssetBucketRef{
					Name: clusterBucketName,
				},
				Source: v1alpha2.AssetSource{
					URL:  assetData.url,
					Mode: assetData.mode,
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

func (a *clusterAsset) delete() error {
	err := a.resCli.Delete(a.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterAsset %s", a.name)
	}

	return nil
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
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready ClusterAsset %s", a.name)
	}

	return nil
}

func (a *clusterAsset) waitForRemove(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		_, err := a.get()
		if err == nil {
			return false, nil
		}

		if !apierrors.IsNotFound(err) {
			return false, err
		}

		return true, nil
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for delete ClusterAsset %s", a.name)
	}

	return err
}
