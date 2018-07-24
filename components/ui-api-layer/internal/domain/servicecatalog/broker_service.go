package servicecatalog

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"k8s.io/client-go/tools/cache"
)

type brokerService struct {
	informer cache.SharedIndexInformer
}

func newBrokerService(informer cache.SharedIndexInformer) *brokerService {
	return &brokerService{
		informer: informer,
	}
}

func (svc *brokerService) Find(name string) (*v1beta1.ClusterServiceBroker, error) {
	item, exists, err := svc.informer.GetStore().GetByKey(name)
	if err != nil || !exists {
		return nil, err
	}

	serviceBroker, ok := item.(*v1beta1.ClusterServiceBroker)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServiceBroker", item)
	}

	return serviceBroker, nil
}

func (svc *brokerService) List(pagingParams pager.PagingParams) ([]*v1beta1.ClusterServiceBroker, error) {
	items, err := pager.From(svc.informer.GetStore()).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var serviceBrokers []*v1beta1.ClusterServiceBroker
	for _, item := range items {
		serviceBroker, ok := item.(*v1beta1.ClusterServiceBroker)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServiceBroker", item)
		}
		serviceBrokers = append(serviceBrokers, serviceBroker)
	}

	return serviceBrokers, nil
}
