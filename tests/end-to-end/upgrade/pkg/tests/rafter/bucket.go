package rafter

import (
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/dynamicresource"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
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
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "buckets",
		}, namespace),
		name:      bucketName,
		namespace: namespace,
	}
}

func (b *bucket) create() error {
	bucket := &v1beta1.Bucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Bucket",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name,
			Namespace: b.namespace,
		},
		Spec: v1beta1.BucketSpec{
			CommonBucketSpec: v1beta1.CommonBucketSpec{
				Policy: v1beta1.BucketPolicyReadOnly,
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

func (b *bucket) get() (*v1beta1.Bucket, error) {
	u, err := b.resCli.Get(b.name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.Bucket
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

		if res.Status.Phase != v1beta1.BucketReady {
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
