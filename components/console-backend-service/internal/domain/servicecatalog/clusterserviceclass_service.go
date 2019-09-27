package servicecatalog

import (
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

type clusterServiceClassService struct {
	informer cache.SharedIndexInformer
}

func newClusterServiceClassService(informer cache.SharedIndexInformer) (*clusterServiceClassService, error) {
	err := informer.AddIndexers(cache.Indexers{
		"externalName": func(obj interface{}) ([]string, error) {
			entity, ok := obj.(*v1beta1.ClusterServiceClass)
			if !ok {
				return nil, errors.New("Cannot convert item")
			}

			return []string{entity.Spec.ExternalName}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	return &clusterServiceClassService{
		informer: informer,
	}, nil
}

func (svc *clusterServiceClassService) Find(name string) (*v1beta1.ClusterServiceClass, error) {
	item, exists, err := svc.informer.GetStore().GetByKey(name)
	if err != nil || !exists {
		return nil, err
	}

	serviceClass, ok := item.(*v1beta1.ClusterServiceClass)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServiceClass", item)
	}

	return serviceClass, nil
}

func (svc *clusterServiceClassService) FindByExternalName(externalName string) (*v1beta1.ClusterServiceClass, error) {
	items, err := svc.informer.GetIndexer().ByIndex("externalName", externalName)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	if len(items) > 1 {
		return nil, fmt.Errorf("Multiple ServiceClass resources with the same externalName %s", externalName)
	}

	item := items[0]
	serviceClass, ok := item.(*v1beta1.ClusterServiceClass)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServiceClass", item)
	}

	return serviceClass, nil
}

func (svc *clusterServiceClassService) List(pagingParams pager.PagingParams) ([]*v1beta1.ClusterServiceClass, error) {
	items, err := pager.From(svc.informer.GetStore()).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var serviceClasses []*v1beta1.ClusterServiceClass
	for _, item := range items {
		serviceClass, ok := item.(*v1beta1.ClusterServiceClass)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServiceClass", item)
		}

		serviceClasses = append(serviceClasses, serviceClass)
	}

	return serviceClasses, nil
}
