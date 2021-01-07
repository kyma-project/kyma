package namespace

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sretry "k8s.io/client-go/util/retry"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/retry"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

const (
	TestNamespaceLabelKey   = "created-by"
	TestNamespaceLabelValue = "serverless-controller-manager-test"
)

type Namespace struct {
	coreCli typedcorev1.CoreV1Interface
	name    string
	log     *logrus.Entry
	verbose bool
}

func New(coreCli typedcorev1.CoreV1Interface, container shared.Container) *Namespace {
	return &Namespace{coreCli: coreCli, name: container.Namespace, log: container.Log, verbose: container.Verbose}
}

func (n Namespace) GetName() string {
	return n.name
}

func (n Namespace) LogResource() error {
	ns, err := n.get()
	if err != nil {
		return err
	}
	ns.TypeMeta.Kind = "namespace"
	out, err := helpers.PrettyMarshall(ns)
	if err != nil {
		return err
	}

	n.log.Infof("%s", out)
	return nil
}

func (n Namespace) get() (*corev1.Namespace, error) {
	return n.coreCli.Namespaces().Get(context.Background(), n.name, metav1.GetOptions{})
}

func (n *Namespace) Create() (string, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n.name,
			Labels: map[string]string{
				"knative-eventing-injection": "enabled",               // https://knative.dev/v0.12-docs/eventing/broker-trigger/#annotation
				TestNamespaceLabelKey:        TestNamespaceLabelValue, // convenience for cleaning up stale namespaces during development
			},
		},
	}

	backoff := wait.Backoff{
		Duration: 500 * time.Millisecond,
		Factor:   2,
		Jitter:   0.1,
		Steps:    4,
	}

	err := k8sretry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		_, err := n.coreCli.Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
		if apierrors.IsAlreadyExists(err) {
			return nil
		}

		return err
	})
	if err != nil {
		return n.name, errors.Wrapf(err, "while creating namespace %s", n.name)
	}

	n.log.Infof("Create: namespace %s", n.name)
	if n.verbose {
		n.log.Infof("%+v", ns)
	}
	return n.name, nil
}

func (n *Namespace) Delete() error {
	err := retry.WithIgnoreOnNotFound(retry.DefaultBackoff, func() error {
		if n.verbose {
			n.log.Infof("DELETE: namespace: %s", n.name)
		}
		return n.coreCli.Namespaces().Delete(context.Background(), n.name, metav1.DeleteOptions{})
	}, n.log)
	if err != nil {
		return errors.Wrapf(err, "while deleting namespace %s", n.name)
	}
	return nil
}
