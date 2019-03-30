package asset_store

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/resource"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"k8s.io/apimachinery/pkg/runtime"
)

type clusterBucket struct {
	resCli      *resource.Resource
	name        string
}

func newClusterBucketClient(dynamicCli dynamic.Interface) *clusterBucket {
	return &clusterBucket{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "clusterbuckets",
		}, ""),
		name: ClusterBucketName,
	}
}

func (b *clusterBucket) create() error {
	clusterBucket := &v1alpha2.ClusterBucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterBucket",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: b.name,
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

func (b *clusterBucket) get() (*v1alpha2.ClusterBucket, error) {
	u, err := b.resCli.Get(b.name)
	if err != nil {
		return nil, err
	}

	var res v1alpha2.ClusterBucket
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterBucket %s", b.name)
	}

	return &res, nil
}

func (b *clusterBucket) waitForStatusReady(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := b.get()
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1alpha2.BucketReady {
			return false, nil
		}

		return true, nil
	}, WaitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready ClusterBucket resource")
	}

	return err
}