package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type bucket struct {
	dynamicCli dynamic.Interface
	res        *resource.Resource
	name string
	namespace string
}

func newBucket(dynamicCli dynamic.Interface, name, namespace string) *bucket {
	return &bucket{
		res: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "buckets",
		}, namespace),
		dynamicCli: dynamicCli,
		name: name,
		namespace:namespace,
	}
}

func (b *bucket) Create() error {
	bucket := &v1alpha2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.name,
			Namespace: b.namespace,
		},
		Spec:v1alpha2.BucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy:v1alpha2.BucketPolicyReadOnly,
			},
		},
	}

	err := b.res.Create(bucket)
	if err != nil {
		return errors.Wrapf(err, "while creating bucket %s in namespace %s", b.name, b.namespace)
	}

	return err
}

func (b *bucket) Delete() error {
	err := b.res.Delete(b.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting bucket %s in namespace %s", b.name, b.namespace)
	}

}
