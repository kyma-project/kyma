package controllers

import (
	"context"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterBucketService struct {
	client client.Client
	scheme *runtime.Scheme
	region string
}

func newClusterBucketService(client client.Client, scheme *runtime.Scheme, region string) *clusterBucketService {
	return &clusterBucketService{
		client: client,
		scheme: scheme,
		region: region,
	}
}

func (s *clusterBucketService) List(ctx context.Context, namespace string, labels map[string]string) ([]string, error) {
	instances := &v1alpha2.ClusterBucketList{}
	err := s.client.List(ctx, instances, client.MatchingLabels(labels))
	if err != nil {
		return nil, errors.Wrapf(err, "while listing ClusterBuckets")
	}

	names := make([]string, 0, len(instances.Items))
	for _, instance := range instances.Items {
		names = append(names, instance.Name)
	}

	return names, nil
}

func (s *clusterBucketService) Create(ctx context.Context, namespacedName types.NamespacedName, private bool, labels map[string]string) error {
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

	if err := s.client.Create(ctx, instance); err != nil {
		return errors.Wrapf(err, "while creating ClusterBucket %s", namespacedName.Name)
	}

	return nil
}
