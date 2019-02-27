package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/waiter"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"time"
)

type clusterAsset struct {
	resCli            *resource.Resource
	ClusterBucketName string
	waitTimeout       time.Duration
}

func newClusterAsset(dynamicCli dynamic.Interface, waitTimeout time.Duration) *clusterAsset {
	return &clusterAsset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "clusterassets",
		}, ""),
		waitTimeout: waitTimeout,
	}
}

func (a *clusterAsset) CreateMany(assets []assetData) error {
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
					BucketRef: v1alpha2.AssetBucketRef{
						Name: a.ClusterBucketName,
					},
					Source:v1alpha2.AssetSource{
						Url: asset.URL,
						Mode:asset.Mode,
					},
				},
			},
		}

		err := a.resCli.Create(asset)
		if err != nil {
			return errors.Wrapf(err, "while creating ClusterAsset %s", asset.Name)
		}
	}

	return nil
}

func (a *clusterAsset) WaitForStatusesReady(assets []assetData) error {
	err := waiter.WaitAtMost(func() (bool, error) {

		for _, asset := range assets {
			res, err := a.Get(asset.Name)
			if err != nil {
				return false, err
			}

			if res.Status.Phase != v1alpha2.AssetReady {
				return false, err
			}
		}

		return true, nil
	}, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready ClusterAsset resources")
	}

	return nil
}

func (a *clusterAsset) VerifyUploadedAssets(assets []assetData, shouldExist bool) error {
	for _, asset := range assets {
		res, err := a.Get(asset.Name)
		if err != nil {
			return err
		}

		err = verifyUploadedAsset("ClusterAsset", asset.Name, res.Status.AssetRef, shouldExist)
		if err != nil {
			return err
		}
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

func (a *clusterAsset) DeleteMany(assets []assetData) error {
	for _, asset := range assets {
		err := a.resCli.Delete(asset.Name)
		if err != nil {
			return errors.Wrapf(err, "while deleting ClusterAsset %s", asset.Name)
		}
	}

	return nil
}
