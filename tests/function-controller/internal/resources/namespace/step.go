package namespace

import (
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	"github.com/pkg/errors"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type NamespaceStep struct {
	ns   *Namespace
	name string
}

func (n NamespaceStep) Name() string {
	return n.name
}

func (n NamespaceStep) Run() error {
	_, err := n.ns.Create()
	return errors.Wrap(err, "while creating namespace")
}

func (n NamespaceStep) Cleanup() error {
	return errors.Wrap(n.ns.Delete(), "while deleting namespace")
}

func (n NamespaceStep) OnError() error {
	return n.ns.LogResource()
}

func NewNamespaceStep(stepName string, coreCli typedcorev1.CoreV1Interface, container utils.Container) NamespaceStep {
	ns := New(coreCli, container)
	return NamespaceStep{name: stepName, ns: ns}
}

var _ executor.Step = NamespaceStep{}
