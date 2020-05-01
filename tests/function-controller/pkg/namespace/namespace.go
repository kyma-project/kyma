package namespace

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/retry"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

const (
	TestNamespaceLabelKey   = "created-by"
	TestNamespaceLabelValue = "serverless-controller-manager-test"
)

type Namespace struct {
	coreCli corev1.CoreV1Interface
	Name    string
	log     shared.Logger
	verbose bool
}

func New(coreCli corev1.CoreV1Interface, name string, log shared.Logger, verbose bool) *Namespace {
	return &Namespace{coreCli: coreCli, Name: name, log: log, verbose: verbose}
}

func (n *Namespace) Create() (string, error) {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n.Name,
			Labels: map[string]string{
				eventingv1alpha1.InjectionAnnotation: "enabled", // https://knative.dev/v0.12-docs/eventing/broker-trigger/#annotation
				TestNamespaceLabelKey:                TestNamespaceLabelValue,
			},
		},
	}

	err := retry.WithIgnoreOnAlreadyExist(retry.DefaultBackoff, func() error {
		_, err := n.coreCli.Namespaces().Create(ns)
		return err
	}, n.log)
	if err != nil {
		return n.Name, errors.Wrapf(err, "while creating namespace %s", n.Name)
	}

	n.log.Logf("CREATE: namespace %s", n.Name)
	if n.verbose {
		n.log.Logf("%v", ns)
	}
	return n.Name, nil
}

func (n *Namespace) Delete() error {
	err := retry.WithIgnoreOnNotFound(retry.DefaultBackoff, func() error {
		n.log.Logf("DELETE: namespace: %s", n.Name)
		return n.coreCli.Namespaces().Delete(n.Name, &metav1.DeleteOptions{})
	}, n.log)
	if err != nil {
		return errors.Wrapf(err, "while deleting namespace %s", n.Name)
	}
	return nil
}
