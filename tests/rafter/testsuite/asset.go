package testsuite

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/tests/rafter/pkg/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type asset struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func newAsset(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *asset {
	return &asset{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assets",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (a *asset) Create(assetData assetData, bucketName string, callbacks ...func(...interface{})) (string, error) {
	asset := &v1beta1.Asset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Asset",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.name,
			Namespace: a.namespace,
		},
		Spec: v1beta1.AssetSpec{
			CommonAssetSpec: v1beta1.CommonAssetSpec{
				BucketRef: v1beta1.AssetBucketRef{
					Name: bucketName,
				},
				Source: v1beta1.AssetSource{
					Mode: assetData.Mode,
					URL:  assetData.URL,
				},
			},
		},
	}

	resourceVersion, err := a.resCli.Create(asset, callbacks...)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating Asset %s in namespace %s", a.name, a.namespace)
	}
	return resourceVersion, err
}

func (a *asset) WaitForStatusReady(initialResourceVersion string, callbacks ...func(...interface{})) error {
	waitForStatusReady := buildWaitForStatusesReady(a.resCli.ResCli, a.waitTimeout, a.name)
	err := waitForStatusReady(initialResourceVersion, callbacks...)
	return err
}

func (a *asset) Get() (*v1beta1.Asset, error) {
	u, err := a.resCli.Get(a.name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.Asset
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Asset %s", a.name)
	}

	return &res, nil
}

func (a *asset) Files() ([]*file, error) {
	asset, err := a.Get()
	if err != nil {
		return nil, errors.Wrapf(err, "while gathering files from Asset %s", a.name)
	}
	statusRef := asset.Status.AssetRef

	var files []*file
	for _, f := range statusRef.Files {
		files = append(files, &file{
			URL:      fmt.Sprintf("%s/%s", statusRef.BaseURL, f.Name),
			Metadata: f.Metadata,
		})
	}
	return files, nil
}

func (a *asset) Delete(callbacks ...func(...interface{})) error {
	err := a.resCli.Delete(a.name, a.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting Asset %s in namespace %s", a.name, a.namespace)
	}

	return nil
}
