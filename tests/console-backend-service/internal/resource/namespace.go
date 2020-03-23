package resource

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Namespace struct {
	resCli *Resource
}

func NewNamespace(dynamicCli dynamic.Interface, logFn func(format string, args ...interface{})) *Namespace {
	return &Namespace{
		resCli: New(dynamicCli, schema.GroupVersionResource{
			Version:  "v1",
			Resource: "namespaces",
		}, "", logFn),
	}
}

func (self *Namespace) Create(name string, labels map[string]string) error {
	Namespace := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	err := self.resCli.Create(Namespace)
	if err != nil {
		return errors.Wrapf(err, "while creating Namespace %s", name)
	}

	return err
}

func (self *Namespace) Delete(name string) error {
	err := self.resCli.Delete(name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Namespace %s", name)
	}

	return nil
}
