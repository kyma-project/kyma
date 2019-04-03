package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/waiter"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type clusterBucket struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func newClusterBucket(dynamicCli dynamic.Interface, name string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *clusterBucket {
	return &clusterBucket{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "clusterbuckets",
		}, "", logFn),
		name:        name,
		waitTimeout: waitTimeout,
	}
}

func (b *clusterBucket) Create() error {
	clusterBucket := &v1alpha2.ClusterBucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterBucket",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name,
			Namespace: b.namespace,
		},
		Spec: v1alpha2.ClusterBucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy: v1alpha2.BucketPolicyReadOnly,
			},
		},
	}

	err := b.resCli.Create(clusterBucket)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterBucket %s", b.name)
	}

	return nil
}

func (b *clusterBucket) WaitForStatusReady() error {
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
		return errors.Wrapf(err, "while waiting for ready ClusterBucket resources")
	}

	return err
}

func (b *clusterBucket) Get(name string) (*v1alpha2.ClusterBucket, error) {
	u, err := b.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1alpha2.ClusterBucket
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterBucket %s", name)
	}

	return &res, nil
}

func (b *clusterBucket) Delete() error {
	err := b.resCli.Delete(b.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterBucket %s", b.name)
	}

	return nil
}
