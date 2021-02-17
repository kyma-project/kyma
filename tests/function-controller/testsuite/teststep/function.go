package teststep

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type newFunction struct {
	name     string
	funcData *function.FunctionData
	fn       *function.Function
	log      *logrus.Entry
}

func CreateFunction(log *logrus.Entry, fn *function.Function, name string, data *function.FunctionData) step.Step {
	return newFunction{
		fn:       fn,
		name:     name,
		funcData: data,
		log:      log.WithField(step.LogStepKey, name),
	}
}

func (f newFunction) Name() string {
	return f.name
}

func (f newFunction) Run() error {
	if err := f.fn.Create(f.funcData); err != nil {
		return errors.Wrapf(err, "while creating function: %s", f.name)
	}

	f.log.Infof("Function Created, Waiting for ready status")
	return errors.Wrapf(f.fn.WaitForStatusRunning(), "while waiting for function: %s, to be ready:", f.name)
}

func (f newFunction) Cleanup() error {
	return errors.Wrapf(f.fn.Delete(), "while deleting function: %s", f.name)
}

func (f newFunction) OnError() error {
	return f.fn.LogResource()
}

var _ step.Step = newFunction{}

type emptyFunction struct {
	name string
	fn   *function.Function
}

func CreateEmptyFunction(fn *function.Function) step.Step {
	return &emptyFunction{
		name: "Creating function without body should be rejected by the webhook",
		fn:   fn,
	}
}

func (e emptyFunction) Name() string {
	return e.name
}

func (e emptyFunction) Run() error {
	err := e.fn.Create(&function.FunctionData{})
	if err == nil {
		return errors.New("Creating empty function should return error, but got nil")
	}
	return nil
}

func (e emptyFunction) Cleanup() error {
	return nil
}

func (e emptyFunction) OnError() error {
	return nil
}

var _ step.Step = emptyFunction{}

type updateFunc struct {
	name     string
	funcData *function.FunctionData
	fn       *function.Function
	log      *logrus.Entry
}

func UpdateFunction(log *logrus.Entry, fn *function.Function, name string, data *function.FunctionData) step.Step {
	return updateFunc{
		fn:       fn,
		name:     name,
		funcData: data,
		log:      log.WithField(step.LogStepKey, name),
	}
}

func (u updateFunc) Name() string {
	return u.name
}

func (u updateFunc) Run() error {
	if err := u.fn.Update(u.funcData); err != nil {
		return errors.Wrapf(err, "while updating function: %s", u.name)
	}

	return errors.Wrapf(u.fn.WaitForStatusRunning(), "while waiting for status ready function: %s", u.name)
}

func (u updateFunc) Cleanup() error {
	return nil
}

func (u updateFunc) OnError() error {
	return nil
}

var _ step.Step = updateFunc{}
