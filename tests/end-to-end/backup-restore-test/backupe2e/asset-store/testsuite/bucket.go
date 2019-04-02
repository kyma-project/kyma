package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/waiter"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type bucket struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func newBucket(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration) *bucket {
	return &bucket{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "buckets",
		}, namespace),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (b *bucket) Create() error {
	bucket := &v1alpha2.Bucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Bucket",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name,
			Namespace: b.namespace,
		},
		Spec: v1alpha2.BucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy: v1alpha2.BucketPolicyReadOnly,
			},
		},
	}

	err := b.resCli.Create(bucket)
	if err != nil {
		return errors.Wrapf(err, "while creating bucket %s in namespace %s", b.name, b.namespace)
	}

	return err
}

func (b *bucket) WaitForStatusReady() error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := b.Get(b.name)
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1alpha2.BucketReady {
			return false, nil
		}

		return true, nil
	}, b.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Bucket resource")
	}

	return nil
}

func (b *bucket) Get(name string) (*v1alpha2.Bucket, error) {
	u, err := b.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1alpha2.Bucket
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Bucket %s", name)
	}

	return &res, nil
}
