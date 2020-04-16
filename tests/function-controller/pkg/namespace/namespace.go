package namespace

import (
	"github.com/kyma-project/kyma/tests/function-controller/pkg/retry"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Namespace struct {
	coreCli corev1.CoreV1Interface
	name    string
	log     shared.Logger
}

func New(coreCli corev1.CoreV1Interface, name string, log shared.Logger) *Namespace {
	return &Namespace{coreCli: coreCli, name: name, log: log}
}

func (n *Namespace) Create() (string, error) {
	err := retry.WithIgnoreOnAlreadyExist(retry.DefaultBackoff, func() error {
		_, err := n.coreCli.Namespaces().Create(&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: n.name,
			},
		})
		return err
	}, n.log)
	if err != nil {
		return n.name, errors.Wrapf(err, "while creating namespace %s", n.name)
	}
	return n.name, nil
}

func (n *Namespace) Delete() error {
	err := retry.WithIgnoreOnNotFound(retry.DefaultBackoff, func() error {
		n.log.Logf("DELETE: namespace: %s", n.name)
		return n.coreCli.Namespaces().Delete(n.name, &metav1.DeleteOptions{})
	}, n.log)
	if err != nil {
		return errors.Wrapf(err, "while deleting namespace %s", n.name)
	}
	return nil
}
