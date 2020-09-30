package scenarios

import (
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

func TestSteps(r *rest.Config, ts testsuite.Config, logf *logrus.Logger) ([]step.Step, error) {

	logger := logf.WithField("", "")
	return []step.Step{
		EmptyStep{msg: "first step", logf: logf},
		EmptyStep{msg: "second step", logf: logf},
		teststep.Parallel(
			teststep.NewSerialTestRunner(logger, "SubTest1",
				EmptyStep{msg: "second step 1", logf: logf},
				EmptyStep{msg: "second step 2", logf: logf},
				EmptyStep{msg: "second step 3", logf: logf},
				EmptyStep{msg: "second step 4", logf: logf, err: errors.New("Error Attention")},
			), teststep.NewSerialTestRunner(logger, "SubTest1",
				EmptyStep{msg: "first step 1", logf: logf},
				EmptyStep{msg: "first step 2", logf: logf},
				EmptyStep{msg: "first step 3", logf: logf},
				EmptyStep{msg: "first step 4", logf: logf},
			)),
	}, nil
}

type EmptyStep struct {
	msg  string
	err  error
	logf *logrus.Logger
}

func (e EmptyStep) Name() string {
	return e.msg
}

func (e EmptyStep) Run() error {
	e.logf.Info(fmt.Sprintf("Run: %s", e.msg))
	return e.err
}

func (e EmptyStep) Cleanup() error {
	return nil
}

func (e EmptyStep) OnError(cause error) error {
	e.logf.Error(fmt.Sprintf("Step: %s, err: %s", e.msg, cause.Error()))
	return nil
}

var _ step.Step = EmptyStep{}
