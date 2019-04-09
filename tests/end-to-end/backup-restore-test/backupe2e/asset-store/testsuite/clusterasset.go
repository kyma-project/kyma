package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/waiter"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/resource"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type clusterAsset struct {
	resCli            *resource.Resource
	clusterBucketName string
	waitTimeout       time.Duration
}

func newClusterAsset(dynamicCli dynamic.Interface, clusterBucketName string, waitTimeout time.Duration) *clusterAsset {
	return &clusterAsset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "clusterassets",
		}, ""),
		waitTimeout:       waitTimeout,
		clusterBucketName: clusterBucketName,
	}
}

func (a *clusterAsset) Create(assetData assetData) error {
	asset := &v1alpha2.ClusterAsset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAsset",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: assetData.Name,
		},
		Spec: v1alpha2.ClusterAssetSpec{
			CommonAssetSpec: v1alpha2.CommonAssetSpec{
				BucketRef: v1alpha2.AssetBucketRef{
					Name: a.clusterBucketName,
				},
				Source: v1alpha2.AssetSource{
					URL:  assetData.URL,
					Mode: assetData.Mode,
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

func (a *clusterAsset) WaitForStatusReady(assetData assetData) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := a.Get(assetData.Name)
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1alpha2.AssetReady {
			return false, nil
		}

		return true, nil
	}, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready ClusterAsset resource")
	}

	return nil
}

func (a *clusterAsset) Get(name string) (*v1alpha2.ClusterAsset, error) {
	u, err := a.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var ca v1alpha2.ClusterAsset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ca)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterAsset %s", name)
	}

	return &ca, nil
}

func (a *clusterAsset) Delete(assetData assetData) error {
	err := a.resCli.Delete(assetData.Name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterAsset %s", assetData.Name)
	}

	return nil
}

func (a *clusterAsset) WaitForDeleted(assetData assetData) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		_, err := a.Get(assetData.Name)
		if err == nil {
			return false, nil
		}

		if !apierrors.IsNotFound(err) {
			return false, err
		}

		return true, nil
	}, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for delete ClusterAsset %s", assetData.Name)
	}

	return err
}
