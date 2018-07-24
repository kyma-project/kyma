package servicecatalog

import (
	"fmt"

	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type serviceBindingService struct {
	client   v1beta1.ServicecatalogV1beta1Interface
	informer cache.SharedIndexInformer
}

func newServiceBindingService(client v1beta1.ServicecatalogV1beta1Interface, informer cache.SharedIndexInformer) *serviceBindingService {
	svc := &serviceBindingService{
		client:   client,
		informer: informer,
	}

	informer.AddIndexers(cache.Indexers{
		"relatedServiceInstanceName": func(obj interface{}) ([]string, error) {
			serviceBinding, err := svc.toServiceBinding(obj)
			if err != nil {
				return nil, errors.Wrapf(err, "while indexing by `relatedServiceInstanceName`")
			}

			key := fmt.Sprintf("%s/%s", serviceBinding.Namespace, serviceBinding.Spec.ServiceInstanceRef.Name)
			return []string{key}, nil
		},
	})

	return svc
}

func (f *serviceBindingService) Create(env string, sb *api.ServiceBinding) (*api.ServiceBinding, error) {
	return f.client.ServiceBindings(env).Create(sb)
}

func (f *serviceBindingService) Delete(env string, name string) error {
	return f.client.ServiceBindings(env).Delete(name, &v1.DeleteOptions{})
}

func (f *serviceBindingService) Find(env string, name string) (*api.ServiceBinding, error) {
	key := fmt.Sprintf("%s/%s", env, name)
	item, exists, err := f.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	return f.toServiceBinding(item)
}

func (f *serviceBindingService) ListForServiceInstance(env string, instanceName string) ([]*api.ServiceBinding, error) {
	key := fmt.Sprintf("%s/%s", env, instanceName)
	items, err := f.informer.GetIndexer().ByIndex("relatedServiceInstanceName", key)
	if err != nil {
		return nil, err
	}

	return f.toServiceBindings(items)
}

func (f *serviceBindingService) toServiceBinding(item interface{}) (*api.ServiceBinding, error) {
	serviceBinding, ok := item.(*api.ServiceBinding)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *ServiceBinding", item)
	}

	return serviceBinding, nil
}

func (f *serviceBindingService) toServiceBindings(items []interface{}) ([]*api.ServiceBinding, error) {
	var serviceBindings []*api.ServiceBinding
	for _, item := range items {
		serviceBinding, err := f.toServiceBinding(item)
		if err != nil {
			return nil, err
		}

		serviceBindings = append(serviceBindings, serviceBinding)
	}

	return serviceBindings, nil
}
