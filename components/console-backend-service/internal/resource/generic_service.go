package resource

import (
	"fmt"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type GenericService struct {
	ServiceBase
}

func (s *GenericService) ListInNamespace(namespace string, result Appendable) error {
	return s.ListByIndex(cache.NamespaceIndex, namespace, result)
}

func (s *GenericService) List(result Appendable) error {
	return s.ListInNamespace("", result)
}

func (s *GenericService) GetInNamespace(name, namespace string, result interface{}) error {
	key := fmt.Sprintf("%s/%s", namespace, name)
	return s.GetByKey(key, result)
}

func (s *GenericService) Get(name string, result interface{}) error {
	return s.GetByKey(name, result)
}

func (s *GenericService) UpdateInNamespace(name, namespace string, result interface{}, update func() error) error {
	err := s.GetInNamespace(name, namespace, result)
	if err != nil {
		return err
	}
	if result == nil {
		return apiErrors.NewNotFound(s.GVR().GroupResource(), name)
	}

	err = update()
	if err != nil {
		return err
	}

	return s.Apply(result, result)
}

func (s *GenericService) DeleteInNamespace(namespace, name string, res runtime.Object) error {
	err := s.GetInNamespace(name, namespace, res)
	if err != nil {
		return err
	}

	res = res.DeepCopyObject()
	err = s.ServiceBase.DeleteInNamespace(name, namespace)
	return err
}

func (s *GenericService) Delete(name string, res runtime.Object) error {
	err := s.Get(name, res)
	if err != nil {
		return err
	}

	res = res.DeepCopyObject()
	err = s.ServiceBase.Delete(name)
	return err
}
