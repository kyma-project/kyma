package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

type podService struct {
	informer cache.SharedIndexInformer
}

func newPodService(informer cache.SharedIndexInformer) *podService {
	return &podService{
		informer: informer,
	}
}

func (svc *podService) Find(name, namespace string) (*v1.Pod, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	pod, ok := item.(*v1.Pod)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *Pod", item)
	}

	return pod, nil
}

func (svc *podService) List(namespace string, pagingParams pager.PagingParams) ([]*v1.Pod, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var pods []*v1.Pod
	for _, item := range items {
		pod, ok := item.(*v1.Pod)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *Pod", item)
		}

		pods = append(pods, pod)
	}

	return pods, nil
}
