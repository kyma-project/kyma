package k8s

import (
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func newLimitRangeService(informer cache.SharedIndexInformer) *limitRangeService {
	return &limitRangeService{
		informer: informer,
	}
}

type limitRangeService struct {
	informer cache.SharedIndexInformer
}

func (svc *limitRangeService) List(env string) ([]*v1.LimitRange, error) {
	items, err := svc.informer.GetIndexer().ByIndex(cache.NamespaceIndex, env)
	if err != nil {
		return []*v1.LimitRange{}, errors.Wrapf(err, "cannot list limit ranges from ns: %s", env)
	}

	var result []*v1.LimitRange
	for _, item := range items {
		lr, ok := item.(*v1.LimitRange)
		if !ok {
			return nil, errors.Errorf("unexpected item type: %T, should be *LimitRange", item)
		}
		result = append(result, lr)
	}

	return result, nil
}
