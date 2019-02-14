package bucket

import (
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type bucketService struct {
	informer cache.SharedIndexInformer
}

//go:generate mockery -name=Lister -output=automock -outpkg=automock -case=underscore
type Lister interface {
	Get(namespace, name string) (*v1alpha1.Bucket, error)
}

func New(informer cache.SharedIndexInformer) *bucketService {
	return &bucketService{
		informer: informer,
	}
}

func (s *bucketService) Get(namespace, name string) (*v1alpha1.Bucket, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := s.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	bucket, ok := item.(*v1alpha1.Bucket)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *Bucket", item)
	}

	return bucket, nil
}
