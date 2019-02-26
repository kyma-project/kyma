package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type clusterBucket struct {
	dynamicCli dynamic.Interface
	res        *resource.Resource
	name string
	namespace string
}

func newClusterBucket(dynamicCli dynamic.Interface, name string) *clusterBucket {
	return &clusterBucket{
		res: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "clusterbuckets",
		}, ""),
		dynamicCli: dynamicCli,
		name: name,
	}
}

func (b *clusterBucket) Create() error {
	clusterBucket := &v1alpha2.ClusterBucket{
		TypeMeta: metav1.TypeMeta{
			Kind: "ClusterBucket",
			APIVersion: v1alpha2.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: b.name,
			Namespace: b.namespace,
		},
		Spec:v1alpha2.ClusterBucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy:v1alpha2.BucketPolicyReadOnly,
			},
		},
	}

	err := b.res.Create(clusterBucket)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterBucket %s in namespace %s", b.name, b.namespace)
	}

	return err
}

func (b *clusterBucket) Delete() error {
	err := b.res.Delete(b.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterBucket %s in namespace %s", b.name, b.namespace)
	}

	return nil
}
