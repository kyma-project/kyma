package k8s

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

type resourceQuotaService struct {
	informer cache.SharedIndexInformer
}

func newResourceQuotaService(informer cache.SharedIndexInformer) *resourceQuotaService {
	return &resourceQuotaService{
		informer: informer,
	}
}

func (svc *resourceQuotaService) List(environment string) ([]*v1.ResourceQuota, error) {
	items, err := svc.informer.GetIndexer().ByIndex(cache.NamespaceIndex, environment)
	if err != nil {
		return []*v1.ResourceQuota{}, err
	}

	var result []*v1.ResourceQuota
	for _, item := range items {
		rq, ok := item.(*v1.ResourceQuota)
		if !ok {
			return nil, fmt.Errorf("unexpected item type: %T, should be *ResourceQuota", item)
		}
		result = append(result, rq)
	}

	return result, nil
}
