package testsuite

import (
	"context"
	"time"

	"k8s.io/client-go/dynamic/dynamicinformer"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/resource"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	watchtools "k8s.io/client-go/tools/watch"
)

type bucket struct {
	resCli          *resource.Resource
	name            string
	namespace       string
	waitTimeout     time.Duration
	informerFactory dynamicinformer.DynamicSharedInformerFactory
}

func newBucket(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *bucket {
	return &bucket{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha2.SchemeGroupVersion.Version,
			Group:    v1alpha2.SchemeGroupVersion.Group,
			Resource: "buckets",
		}, namespace, logFn),
		name:            name,
		namespace:       namespace,
		waitTimeout:     waitTimeout,
		informerFactory: dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamicCli, waitTimeout, namespace, nil),
	}
}

func (b *bucket) Create(callbacks ...func(...interface{})) (string, error) {
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
	resourceVersion, err := b.resCli.Create(bucket, callbacks...)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating Bucket %s in namespace %s", b.name, b.namespace)
	}
	return resourceVersion, err
}

func (b *bucket) WaitForStatusReady(initialResourceVersion string, callbacks ...func(...interface{})) error {
	ctx, cancel := context.WithTimeout(context.Background(), b.waitTimeout)
	defer cancel()
	condition := isPhaseReady(b.name, callbacks...)
	_, err := watchtools.Until(ctx, initialResourceVersion, b.resCli.ResCli, condition)
	if err != nil {
		return err
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

func (b *bucket) Delete(callbacks ...func(...interface{})) error {
	err := b.resCli.Delete(b.name, b.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting Bucket %s in namespace %s", b.name, b.namespace)
	}
	return nil
}
