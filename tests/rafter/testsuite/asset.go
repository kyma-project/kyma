package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/tests/rafter/pkg/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
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
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assets",
		}, namespace, logFn),
		waitTimeout: waitTimeout,
		BucketName:  bucketName,
		Namespace:   namespace,
	}
}

func (a *asset) CreateMany(assets []assetData, testID string, callbacks ...func(...interface{})) (string, error) {
	var initialResourceVersion string
	for _, asset := range assets {
		asset := &v1beta1.Asset{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Asset",
				APIVersion: v1beta1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      asset.Name,
				Namespace: a.Namespace,
				Labels: map[string]string{
					"test-id": testID,
				},
			},
			Spec: v1beta1.AssetSpec{
				CommonAssetSpec: v1beta1.CommonAssetSpec{
					BucketRef: v1beta1.AssetBucketRef{
						Name: a.BucketName,
					},
					Source: v1beta1.AssetSource{
						URL:  asset.URL,
						Mode: asset.Mode,
					},
				},
			},
		}
		resourceVersion, err := a.resCli.Create(asset, callbacks...)
		if err != nil {
			return initialResourceVersion, errors.Wrapf(err, "while creating Asset %s in namespace %s", asset.Name, a.Namespace)
		}
		if initialResourceVersion != "" {
			continue
		}
		initialResourceVersion = resourceVersion
	}
	return initialResourceVersion, nil
}

func (a *asset) WaitForStatusesReady(assets []assetData, resourceVersion string, callbacks ...func(...interface{})) error {
	var assetNames []string
	for _, asset := range assets {
		assetNames = append(assetNames, asset.Name)
	}
	waitForStatusesReady := buildWaitForStatusesReady(a.resCli.ResCli, a.waitTimeout, assetNames...)
	err := waitForStatusesReady(resourceVersion, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while waiting for assets to have ready state")
	}
	return nil
}

func (a *asset) PopulateUploadFiles(assets []assetData, callbacks ...func(...interface{})) ([]uploadedFile, error) {
	var files []uploadedFile

	for _, asset := range assets {
		res, err := a.Get(asset.Name, callbacks...)
		if err != nil {
			return nil, err
		}
		assetFiles := uploadedFiles(res.Status.CommonAssetStatus.AssetRef, res.Name, "Asset")
		files = append(files, assetFiles...)
	}

	return files, nil
}

func (a *asset) Get(name string, callbacks ...func(...interface{})) (*v1beta1.Asset, error) {
	u, err := a.resCli.Get(name, callbacks...)
	if err != nil {
		return nil, err
	}

	var res v1beta1.Asset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while converting Asset %s", name)
	}

	return &res, nil
}

func (a *asset) DeleteLeftovers(testId string, callbacks ...func(...interface{})) error {
	deleteLeftovers := buildDeleteLeftovers(a.resCli.ResCli, a.waitTimeout)
	err := deleteLeftovers(testId, callbacks...)
	return err
}
