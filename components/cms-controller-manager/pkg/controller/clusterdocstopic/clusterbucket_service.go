package clusterdocstopic

import (
	"context"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type bucketService struct {
	client client.Client
	scheme *runtime.Scheme
	region string
}

func newClusterBucketService(client client.Client, scheme *runtime.Scheme, region string) *bucketService {
	return &bucketService{
		client: client,
		scheme: scheme,
		region: region,
	}
}

func (s *bucketService) List(ctx context.Context, labels map[string]string) ([]types.NamespacedName, error) {
	instances := &v1alpha2.ClusterBucketList{}
	err := s.client.List(ctx, client.MatchingLabels(labels), instances)
	if err != nil {
		return nil, err
	}

	namespacedNames := make([]types.NamespacedName, 0, len(instances.Items))
	for _, instance := range instances.Items {
		namespacedName := types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}
		namespacedNames = append(namespacedNames, namespacedName)
	}

	return namespacedNames, nil
}

func (s *bucketService) Create(ctx context.Context, namespacedName types.NamespacedName, private bool, labels map[string]string) error {
	policy := v1alpha2.BucketPolicyReadOnly
	if private {
		policy = v1alpha2.BucketPolicyNone
	}

	instance := &v1alpha2.ClusterBucket{
		ObjectMeta: v1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
			Labels:    labels,
		},
		Spec: v1alpha2.ClusterBucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy: policy,
				Region: v1alpha2.BucketRegion(s.region),
			},
		},
	}

	return s.client.Create(ctx, instance)
}
