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

type clusterBucket struct {
	resCli *dynamicresource.DynamicResource
	name   string
}

func newClusterBucket(dynamicCli dynamic.Interface) *clusterBucket {
	return &clusterBucket{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "clusterbuckets",
		}, ""),
		name: clusterBucketName,
	}
}

func (b *clusterBucket) create() error {
	clusterBucket := &v1beta1.ClusterBucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterBucket",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: b.name,
		},
		Spec: v1beta1.ClusterBucketSpec{
			CommonBucketSpec: v1beta1.CommonBucketSpec{
				Policy: v1beta1.BucketPolicyReadOnly,
			},
		},
	}

	err := b.resCli.Create(clusterBucket)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterBucket %s", b.name)
	}

	return nil
}

func (b *clusterBucket) get() (*v1beta1.ClusterBucket, error) {
	u, err := b.resCli.Get(b.name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.ClusterBucket
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterBucket %s", b.name)
	}

	return &res, nil
}

func (b *clusterBucket) delete() error {
	err := b.resCli.Delete(b.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterBucket %s", b.name)
	}

	return nil
}

func (b *clusterBucket) waitForStatusReady(stop <-chan struct{}) error {
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
		return errors.Wrapf(err, "while waiting for ready ClusterBucket %s", b.name)
	}

	return err
}

func (b *clusterBucket) waitForRemove(stop <-chan struct{}) error {
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
		return errors.Wrapf(err, "while waiting for delete ClusterBucket %s", b.name)
	}

	return err
}
