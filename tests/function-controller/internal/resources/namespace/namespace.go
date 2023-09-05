package namespace

import (
	"context"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8sretry "k8s.io/client-go/util/retry"
)

const (
	TestNamespaceLabelKey   = "created-by"
	TestNamespaceLabelValue = "serverless-controller-manager-test"
)

type Namespace struct {
	coreCli   typedcorev1.CoreV1Interface
	namespace string
	log       *logrus.Entry
}

func New(log *logrus.Entry, coreCli typedcorev1.CoreV1Interface, namespace string) *Namespace {
	return &Namespace{coreCli: coreCli, namespace: namespace, log: log}
}

func (n Namespace) LogResource() error {
	ns, err := n.get()
	if err != nil {
		return err
	}
	ns.TypeMeta.Kind = "namespace"
	out, err := utils.PrettyMarshall(ns)
	if err != nil {
		return err
	}

	n.log.Infof("%s", out)
	return nil
}

func (n Namespace) get() (*corev1.Namespace, error) {
	return n.coreCli.Namespaces().Get(context.Background(), n.namespace, metav1.GetOptions{})
}

func (n *Namespace) Create() (string, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: n.namespace,
			Labels: map[string]string{
				TestNamespaceLabelKey: TestNamespaceLabelValue, // convenience for cleaning up stale namespaces during development
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
		return n.namespace, errors.Wrapf(err, "while creating namespace %s", n.namespace)
	}

	n.log.Infof("Create: namespace %s", n.namespace)
	return n.namespace, nil
}

func (n *Namespace) Delete() error {
	err := utils.WithIgnoreOnNotFound(utils.DefaultBackoff, func() error {
		return n.coreCli.Namespaces().Delete(context.Background(), n.namespace, metav1.DeleteOptions{})
	}, n.log)
	if err != nil {
		return errors.Wrapf(err, "while deleting namespace %s", n.namespace)
	}
	return nil
}
