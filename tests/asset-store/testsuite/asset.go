package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/waiter"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type asset struct {
	resCli      *resource.Resource
	BucketName  string
	Namespace   string
	waitTimeout time.Duration
}

func newAsset(dynamicCli dynamic.Interface, namespace string, bucketName string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *asset {
	return &asset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "assets",
		}, namespace, logFn),
		waitTimeout: waitTimeout,
		BucketName:  bucketName,
		Namespace:   namespace,
	}
}

func (a *asset) CreateMany(assets []assetData) error {
	for _, asset := range assets {
		asset := &v1alpha2.Asset{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Asset",
				APIVersion: v1alpha2.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      asset.Name,
				Namespace: a.Namespace,
			},
			Spec: v1alpha2.AssetSpec{
				CommonAssetSpec: v1alpha2.CommonAssetSpec{
					BucketRef: v1alpha2.AssetBucketRef{
						Name: a.BucketName,
					},
					Source: v1alpha2.AssetSource{
						Url:  asset.URL,
						Mode: asset.Mode,
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

func (a *asset) WaitForStatusesReady(assets []assetData) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		for _, asset := range assets {
			res, err := a.Get(asset.Name)
			if err != nil {
				return false, err
			}

			if res.Status.Phase != v1alpha2.AssetReady {
				return false, nil
			}
		}

		return true, nil
	}, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Asset resources")
	}

	return nil
}

func (a *asset) WaitForDeletedResources(assets []assetData) error {
	err := waiter.WaitAtMost(func() (bool, error) {

		for _, asset := range assets {
			_, err := a.Get(asset.Name)
			if err == nil {
				return false, nil
			}

			if !apierrors.IsNotFound(err) {
				return false, nil
			}
		}

		return true, nil
	}, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready ClusterAsset resources")
	}

	return nil
}

func (a *asset) PopulateUploadFiles(assets []assetData) ([]uploadedFile, error) {
	var files []uploadedFile

	for _, asset := range assets {
		res, err := a.Get(asset.Name)
		if err != nil {
			return nil, err
		}

		assetFiles := uploadedFiles(res.Status.CommonAssetStatus.AssetRef, res.Name, "Asset")
		files = append(files, assetFiles...)
	}

	return files, nil
}

func (a *asset) Get(name string) (*v1alpha2.Asset, error) {
	u, err := a.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1alpha2.Asset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while converting Asset %s", name)
	}

	return &res, nil
}

func (a *asset) DeleteMany(assets []assetData) error {
	for _, asset := range assets {
		err := a.resCli.Delete(asset.Name)
		if err != nil {
			return errors.Wrapf(err, "while deleting Asset %s in namespace %s", asset.Name, a.Namespace)
		}
	}

	return nil
}
