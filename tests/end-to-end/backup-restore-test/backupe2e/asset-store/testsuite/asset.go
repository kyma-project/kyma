package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/resource"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/waiter"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type asset struct {
	resCli      *resource.Resource
	bucketName  string
	namespace   string
	waitTimeout time.Duration
}

func newAsset(dynamicCli dynamic.Interface, bucketName, namespace string, waitTimeout time.Duration) *asset {
	return &asset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.GroupVersion.Version,
			Group:    v1alpha2.GroupVersion.Group,
			Resource: "assets",
		}, namespace),
		waitTimeout: waitTimeout,
		bucketName:  bucketName,
		namespace:   namespace,
	}
}

func (a *asset) Create(assetData assetData) error {
	asset := &v1alpha2.Asset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Asset",
			APIVersion: v1alpha2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      assetData.Name,
			Namespace: a.namespace,
		},
		Spec: v1alpha2.AssetSpec{
			CommonAssetSpec: v1alpha2.CommonAssetSpec{
				BucketRef: v1alpha2.AssetBucketRef{
					Name: a.bucketName,
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
		return errors.Wrapf(err, "while creating Asset %s in namespace %s", asset.Name, a.namespace)
	}

	return nil
}

func (a *asset) WaitForStatusReady(assetData assetData) error {
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
		return errors.Wrapf(err, "while waiting for ready Asset resource")
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
		if apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while converting Asset %s", name)
	}

	return &res, nil
}
