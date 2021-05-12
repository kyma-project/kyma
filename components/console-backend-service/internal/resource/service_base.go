package resource

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

var NotFound = fmt.Errorf("resource not found")

type Appendable interface {
	Append() interface{}
}

type Unsubscribe func()

type ServiceBase interface {
	ListByIndex(index, key string, result Appendable) error
	GetByKey(key string, result interface{}) error
	AddIndexers(indexers cache.Indexers) error
	Create(obj interface{}, result interface{}) error
	Apply(obj interface{}, result interface{}) error
	GVR() schema.GroupVersionResource
	DeleteInNamespace(name, namespace string) error
	Delete(name string) error
	Subscribe(handler EventHandlerProvider) (Unsubscribe, error)
}

type enabledServiceBase struct {
	Client   dynamic.NamespaceableResourceInterface
	Informer cache.SharedIndexInformer
	notifier *Notifier
	gvr      schema.GroupVersionResource
}

func (s *enabledServiceBase) ListByIndex(index, key string, result Appendable) error {
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

func (s *enabledServiceBase) GetByKey(key string, result interface{}) error {
	item, exists, err := s.Informer.GetStore().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return NotFound
	}

	err = FromUnstructured(item.(*unstructured.Unstructured), result)
	if err != nil {
		return err
	}

	return nil
}

func (s *enabledServiceBase) AddIndexers(indexers cache.Indexers) error {
	err := s.Informer.AddIndexers(indexers)
	if err != nil && err.Error() == "informer has already started" {
		return nil
	}
	return err
}

func (s *enabledServiceBase) Create(obj interface{}, result interface{}) error {
	u, err := ToUnstructured(obj)
	if err != nil {
		return err
	}

	var created *unstructured.Unstructured
	if u.GetNamespace() == "" {
		created, err = s.Client.Create(context.Background(), u, v1.CreateOptions{})
	} else {
		created, err = s.Client.Namespace(u.GetNamespace()).Create(context.Background(), u, v1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	return FromUnstructured(created, result)
}

func (s *enabledServiceBase) Apply(obj interface{}, result interface{}) error {
	u, err := ToUnstructured(obj)
	if err != nil {
		return err
	}

	var updated *unstructured.Unstructured
	if u.GetNamespace() == "" {
		updated, err = s.Client.Update(context.Background(), u, v1.UpdateOptions{})
	} else {
		updated, err = s.Client.Namespace(u.GetNamespace()).Update(context.Background(), u, v1.UpdateOptions{})
	}
	if err != nil {
		return err
	}

	return FromUnstructured(updated, result)
}

func (s *enabledServiceBase) Subscribe(handler EventHandlerProvider) (Unsubscribe, error) {
	listener := NewListener(handler)
	s.notifier.AddListener(listener)
	return func() {
		s.deleteListener(listener)
	}, nil
}

func (s *enabledServiceBase) deleteListener(listener *Listener) {
	s.notifier.DeleteListener(listener)
}

func (s *enabledServiceBase) DeleteInNamespace(name, namespace string) error {
	return s.Client.Namespace(namespace).Delete(context.Background(), name, v1.DeleteOptions{})
}

func (s *enabledServiceBase) Delete(name string) error {
	return s.Client.Delete(context.Background(), name, v1.DeleteOptions{})
}

func (s *enabledServiceBase) GVR() schema.GroupVersionResource {
	return s.gvr
}

type disabledServiceBase struct {
	gvr schema.GroupVersionResource
	err error
}

func (s disabledServiceBase) ListByIndex(_, _ string, _ Appendable) error {
	return s.err
}

func (s disabledServiceBase) GetByKey(_ string, _ interface{}) error {
	return s.err
}

func (s disabledServiceBase) AddIndexers(_ cache.Indexers) error {
	return s.err
}

func (s disabledServiceBase) Create(_ interface{}, _ interface{}) error {
	return s.err
}

func (s disabledServiceBase) Apply(_ interface{}, _ interface{}) error {
	return s.err
}

func (s disabledServiceBase) GVR() schema.GroupVersionResource {
	return s.gvr
}

func (s disabledServiceBase) DeleteInNamespace(_, _ string) error {
	return s.err
}

func (s disabledServiceBase) Delete(_ string) error {
	return s.err
}

func (s disabledServiceBase) Subscribe(_ EventHandlerProvider) (Unsubscribe, error) {
	return func() {}, s.err
}
