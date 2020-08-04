package teststep

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/pkg/errors"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
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

func NewNamespaceStep(logger shared.Logger, coreCli typedcorev1.CoreV1Interface, name string) NamespaceStep {
	ns := namespace.New2(name, coreCli, logger)
	return NamespaceStep{name: name, ns: ns}
}

var _ step.Step = NamespaceStep{}
