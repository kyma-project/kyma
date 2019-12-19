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

type bucket struct {
	resCli    *dynamicresource.DynamicResource
	name      string
	namespace string
}

func newBucket(dynamicCli dynamic.Interface, namespace string) *bucket {
	return &bucket{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.GroupVersion.Version,
			Group:    v1alpha2.GroupVersion.Group,
			Resource: "buckets",
		}, namespace),
		name:      bucketName,
		namespace: namespace,
	}
}

func (b *bucket) create() error {
	bucket := &v1alpha2.Bucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Bucket",
			APIVersion: v1alpha2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name,
			Namespace: b.namespace,
		},
		Spec: v1alpha2.BucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy: v1alpha2.BucketPolicyReadOnly,
				Region: bucketRegion,
			},
		},
	}

	err := b.resCli.Create(bucket)
	if err != nil {
		return errors.Wrapf(err, "while creating Bucket %s in namespace %s", b.name, b.namespace)
	}

	return nil
}

func (b *bucket) get() (*v1alpha2.Bucket, error) {
	u, err := b.resCli.Get(b.name)
	if err != nil {
		return nil, err
	}

	var res v1alpha2.Bucket
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Bucket %s in namespace %s", b.name, b.namespace)
	}

	return &res, nil
}

func (b *bucket) delete() error {
	err := b.resCli.Delete(b.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Bucket %s in namespace %s", b.name, b.namespace)
	}

	return nil
}

func (b *bucket) waitForStatusReady(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := b.get()
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1alpha2.BucketReady {
			return false, nil
		}

		return true, nil
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Bucket %s in namespace %s", b.name, b.namespace)
	}

	return err
}

func (b *bucket) waitForRemove(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		_, err := b.get()
		if err == nil {
			return false, nil
		}

		if !apierrors.IsNotFound(err) {
			return false, err
		}

		return true, nil
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for delete Bucket %s in namespace %s", b.name, b.namespace)
	}

	return err
}
