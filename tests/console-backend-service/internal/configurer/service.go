package configurer

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Service struct {
	name      string
	namespace string
	coreCli   *corev1.CoreV1Client
}

func NewService(coreCli *corev1.CoreV1Client, name, namespace string) *Service {
	return &Service{
		coreCli: coreCli, name: name, namespace: namespace,
	}
}

func (self *Service) Create(spec v1.ServiceSpec) error {
	Service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      self.name,
			Namespace: self.namespace,
		},
		Spec: spec,
	}

	_, err := self.coreCli.Services(self.namespace).Create(Service)
	if err != nil {
		return errors.Wrapf(err, "while creating Service %s", self.name)
	}

	return err
}

func (self *Service) Delete() error {
	err := self.coreCli.Services(self.namespace).Delete(self.name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "while deleting Service %s", self.name)
	}

	return nil
}
