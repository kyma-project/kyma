package servicecatalog

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

type serviceClassService struct {
	informer cache.SharedIndexInformer
}

func newServiceClassService(informer cache.SharedIndexInformer) *serviceClassService {
	informer.AddIndexers(cache.Indexers{
		"externalName": func(obj interface{}) ([]string, error) {
			entity, ok := obj.(*v1beta1.ServiceClass)
			if !ok {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Namespace, entity.Spec.ExternalName)}, nil
		},
	})

	return &serviceClassService{
		informer: informer,
	}
}

func (svc *serviceClassService) Find(name, environment string) (*v1beta1.ServiceClass, error) {
	key := fmt.Sprintf("%s/%s", environment, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	serviceClass, ok := item.(*v1beta1.ServiceClass)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServiceClass", item)
	}

	return serviceClass, nil
}

func (svc *serviceClassService) FindByExternalName(externalName, environment string) (*v1beta1.ServiceClass, error) {
	key := fmt.Sprintf("%s/%s", environment, externalName)
	items, err := svc.informer.GetIndexer().ByIndex("externalName", key)
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
	serviceClass, ok := item.(*v1beta1.ServiceClass)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServiceClass", item)
	}

	return serviceClass, nil
}

func (svc *serviceClassService) List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceClass, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", environment).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var serviceClasses []*v1beta1.ServiceClass
	for _, item := range items {
		serviceClass, ok := item.(*v1beta1.ServiceClass)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServiceClass", item)
		}

		serviceClasses = append(serviceClasses, serviceClass)
	}

	return serviceClasses, nil
}
