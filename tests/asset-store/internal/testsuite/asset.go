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

type asset struct {
	resCli     *resource.Resource
	BucketName string
	Namespace  string
	waitTimeout       time.Duration
}

func newAsset(dynamicCli dynamic.Interface, namespace string, waitTimeout time.Duration) *asset {
	return &asset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "assets",
		}, namespace),
		waitTimeout: waitTimeout,
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

		err := a.resCli.Create(asset)
		if err != nil {
			return errors.Wrapf(err, "while creating Asset %s in namespace %s", asset.Name, a.Namespace)
		}
	}

	return nil
}

func (a *asset) WaitForStatusReady(assets []assetData) error {
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
		return errors.Wrapf(err, "while waiting for ready Asset resources")
	}

	return nil
}

func (a *asset) VerifyUploadedAssets(assets []assetData, shouldExist bool) error {
	for _, asset := range assets {
		res, err := a.Get(asset.Name)
		if err != nil {
			return err
		}

		err = verifyUploadedAsset("Asset", asset.Name, res.Status.AssetRef, shouldExist)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *asset) Get(name string) (*v1alpha2.Asset, error) {
	u, err := a.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1alpha2.Asset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Asset %s", name)
	}

	return &res, nil
}

func (a *asset) Delete(assets []assetData) error {
	for _, asset := range assets {
		err := a.resCli.Delete(asset.Name)
		if err != nil {
			return errors.Wrapf(err, "while deleting Asset %s in namespace %s", asset.Name, a.Namespace)
		}
	}

	return nil
}

