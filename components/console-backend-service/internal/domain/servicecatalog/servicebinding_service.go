package servicecatalog

import (
	"context"
	"fmt"

	api "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type serviceBindingService struct {
	client   v1beta1.ServicecatalogV1beta1Interface
	informer cache.SharedIndexInformer
	notifier notifier

	nameFunc func() string
}

func newServiceBindingService(client v1beta1.ServicecatalogV1beta1Interface, informer cache.SharedIndexInformer, nameFunc func() string) (*serviceBindingService, error) {
	svc := &serviceBindingService{
		client:   client,
		informer: informer,
		nameFunc: nameFunc,
	}

	err := informer.AddIndexers(cache.Indexers{
		"relatedServiceInstanceName": func(obj interface{}) ([]string, error) {
			serviceBinding, err := svc.toServiceBinding(obj)
			if err != nil {
				return nil, errors.Wrapf(err, "while indexing by `relatedServiceInstanceName`")
			}

			key := fmt.Sprintf("%s/%s", serviceBinding.Namespace, serviceBinding.Spec.InstanceRef.Name)
			return []string{key}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)

	svc.notifier = notifier

	return svc, nil
}

func (f *serviceBindingService) Create(ns string, sb *api.ServiceBinding) (*api.ServiceBinding, error) {
	if sb.Name == "" {
		sb.Name = f.nameFunc()
	}
	return f.client.ServiceBindings(ns).Create(context.Background(), sb, v1.CreateOptions{})
}

func (f *serviceBindingService) Delete(ns string, name string) error {
	return f.client.ServiceBindings(ns).Delete(context.Background(), name, v1.DeleteOptions{})
}

func (f *serviceBindingService) Find(ns string, name string) (*api.ServiceBinding, error) {
	key := fmt.Sprintf("%s/%s", ns, name)
	item, exists, err := f.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	return f.toServiceBinding(item)
}

func (f *serviceBindingService) ListForServiceInstance(ns string, instanceName string) ([]*api.ServiceBinding, error) {
	key := fmt.Sprintf("%s/%s", ns, instanceName)
	items, err := f.informer.GetIndexer().ByIndex("relatedServiceInstanceName", key)
	if err != nil {
		return nil, err
	}

	return f.toServiceBindings(items)
}

func (f *serviceBindingService) Subscribe(listener resource.Listener) {
	f.notifier.AddListener(listener)
}

func (f *serviceBindingService) Unsubscribe(listener resource.Listener) {
	f.notifier.DeleteListener(listener)
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
