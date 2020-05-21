package resource

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"reflect"
	"time"

	"github.com/pkg/errors"

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

func (s *Service) ListInIndex(index, key string, results interface{}) error {
	resultsVal := reflect.ValueOf(results)
	if resultsVal.Kind() != reflect.Ptr {
		return errors.New("results argument must be a pointer to a slice")
	}

	sliceVal := resultsVal.Elem()
	elementType := sliceVal.Type().Elem()

	items, err := s.Informer.GetIndexer().ByIndex(index, key)
	if err != nil {
		return err
	}

	sliceVal, err = s.addItems(sliceVal, elementType, items)
	if err != nil {
		return err
	}

	resultsVal.Elem().Set(sliceVal)
	return nil
}

func (s *Service) GetInNamespace(name, namespace string, result Convertible) error {
	item, err := s.Client.Namespace(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}

	err = result.FromUnstructured(item)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Get(name string, result Convertible) error {
	item, err := s.Client.Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}

	err = result.FromUnstructured(item)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) List(result ConvertibleList) error {
	list, err := s.Client.List(v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ns := range list.Items {
		converted := result.Append()
		err := converted.FromUnstructured(&ns)
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

func (s *Service) addItems(sliceVal reflect.Value, elemType reflect.Type, items []interface{}) (reflect.Value, error) {
	for index, item := range items {
		if sliceVal.Len() == index {
			// slice is full
			newElem := reflect.New(elemType)
			sliceVal = reflect.Append(sliceVal, newElem.Elem())
			sliceVal = sliceVal.Slice(0, sliceVal.Cap())
		}

		currElem := sliceVal.Index(index).Addr().Interface()
		err := FromUnstructured(item.(*unstructured.Unstructured), currElem)
		if err != nil {
			return sliceVal, err
		}
	}

	return sliceVal, nil
}
