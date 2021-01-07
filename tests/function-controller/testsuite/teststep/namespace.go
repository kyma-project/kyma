package teststep

import (
	"github.com/pkg/errors"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type NamespaceStep struct {
	ns   *namespace.Namespace
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

func NewNamespaceStep(stepName string, coreCli typedcorev1.CoreV1Interface, container shared.Container) NamespaceStep {
	ns := namespace.New(coreCli, container)
	return NamespaceStep{name: stepName, ns: ns}
}

var _ step.Step = NamespaceStep{}
