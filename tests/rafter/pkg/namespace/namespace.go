package namespace

import (
	"fmt"

	"github.com/kyma-project/kyma/tests/rafter/pkg/retry"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Namespace struct {
	coreCli corev1.CoreV1Interface
	name    string
}

func New(coreCli corev1.CoreV1Interface, name string) *Namespace {
	return &Namespace{coreCli: coreCli, name: name}
}

func (n *Namespace) Create(callbacks ...func(...interface{})) error {
	err := retry.WithIgnoreOnAlreadyExist(retry.DefaultBackoff, func() error {
		_, err := n.coreCli.Namespaces().Create(&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: n.name,
			},
		})
		return err
	}, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while creating namespace %s", n.name)
	}
	return nil
}

func (n *Namespace) Delete(callbacks ...func(...interface{})) error {
	err := retry.WithIgnoreOnNotFound(retry.DefaultBackoff, func() error {
		for _, callback := range callbacks {
			callback(fmt.Sprintf("DELETE: namespace: %s", n.name))
		}
		return n.coreCli.Namespaces().Delete(n.name, &metav1.DeleteOptions{})
	}, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting namespace %s", n.name)
	}
	return nil
}
