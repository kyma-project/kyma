package servicecatalog

import (
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"k8s.io/client-go/tools/cache"
)

type clusterServiceBrokerService struct {
	informer cache.SharedIndexInformer
	notifier notifier
}

func newClusterServiceBrokerService(informer cache.SharedIndexInformer) *clusterServiceBrokerService {
	return &clusterServiceBrokerService{
		informer: informer,
		notifier: resource.NewNotifier(),
	}
}

func (svc *clusterServiceBrokerService) Find(name string) (*v1beta1.ClusterServiceBroker, error) {
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

func (svc *clusterServiceBrokerService) List(pagingParams pager.PagingParams) ([]*v1beta1.ClusterServiceBroker, error) {
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

func (svc *clusterServiceBrokerService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *clusterServiceBrokerService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}
