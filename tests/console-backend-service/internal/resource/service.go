package resource

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Service struct {
	resCli *Resource
}

func NewService(dynamicCli dynamic.Interface, namespace string, logFn func(format string, args ...interface{})) *Service {
	return &Service{
		resCli: New(dynamicCli, schema.GroupVersionResource{
			Version:  "v1",
			Resource: "services",
		}, namespace, logFn),
	}
}

func (self *Service) Create(name string, spec v1.ServiceSpec) error {
	Service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: spec,
	}

	err := self.resCli.Create(Service)
	if err != nil {
		return errors.Wrapf(err, "while creating Service %s", name)
	}

	return err
}

func (self *Service) Delete(name string) error {
	err := self.resCli.Delete(name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Service %s", name)
	}

	return nil
}
