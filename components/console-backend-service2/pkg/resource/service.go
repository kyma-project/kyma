package resource

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type ServiceFactory struct {
	Client          dynamic.Interface
	InformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func NewServiceFactoryForConfig(config *rest.Config, informerResyncPeriod time.Duration) (*ServiceFactory, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, informerResyncPeriod)
	return NewServiceFactory(dynamicClient, informerFactory), nil
}

func NewServiceFactory(client dynamic.Interface, informerFactory dynamicinformer.DynamicSharedInformerFactory) *ServiceFactory {
	return &ServiceFactory{
		Client:          client,
		InformerFactory: informerFactory,
	}
}

type Service struct {
	Client   dynamic.NamespaceableResourceInterface
	Informer cache.SharedIndexInformer
	new      func() interface{}
}

func (f *ServiceFactory) ForResource(gvr schema.GroupVersionResource) *Service {
	return &Service{
		Client:   f.Client.Resource(gvr),
		Informer: f.InformerFactory.ForResource(gvr).Informer(),
	}
}

func (s *Service) ListInIndex(index, key string, result Appendable) error {
	items, err := s.Informer.GetIndexer().ByIndex(index, key)
	if err != nil {
		return err
	}

	for _, item := range items {
		converted := result.Append()
		err := FromUnstructured(item.(*unstructured.Unstructured), converted)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetInNamespace(name, namespace string, result interface{}) error {
	item, err := s.Client.Namespace(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}

	err = FromUnstructured(item, result)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Get(name string, result interface{}) error {
	item, err := s.Client.Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}

	err = FromUnstructured(item, result)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) List(result Appendable) error {
	list, err := s.Client.List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		converted := result.Append()
		err := FromUnstructured(&item, converted)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) AddIndexers(indexers cache.Indexers) error {
	err := s.Informer.AddIndexers(indexers)
	if err != nil && err.Error() == "informer has already started" {
		return nil
	}
	return err
}