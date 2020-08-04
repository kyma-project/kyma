package teststep

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/pkg/errors"
)

type newFunction struct {
	name     string
	funcData function.FunctionData
	fn       *function.Function
	log      shared.Logger
}

func CreateFunction(log shared.Logger, fn *function.Function, stepName string, data function.FunctionData) step.Step {
	return newFunction{
		fn:       fn,
		name:     stepName,
		funcData: data,
		log:      log,
	}
}

func (f newFunction) Name() string {
	return f.name
}

func (f newFunction) Run() error {
	if err := f.fn.Create(&f.funcData); err != nil {
		return errors.Wrapf(err, "while creating function: %s", f.name)
	}

	f.log.Logf("Function Created, Waiting for ready status")
	return errors.Wrapf(f.fn.WaitForStatusRunning(), "while waiting for function: %s, to be ready:", f.name)
}

func (f newFunction) Cleanup() error {
	return errors.Wrapf(f.fn.Delete(), "while deleting function: %s", f.name)
}

var _ step.Step = newFunction{}
var _ step.Step = updateFunc{}

type updateFunc struct {
	name     string
	funcData function.FunctionData
	fn       *function.Function
	log      shared.Logger
}

func (u updateFunc) Name() string {
	return u.name
}

func (u updateFunc) Run() error {
	return errors.Wrapf(u.fn.Update(&u.funcData), "while updating function: %s", u.name)
}

func (u updateFunc) Cleanup() error {
	return nil
}

func UpdateFunction(log shared.Logger, fn *function.Function, name string, data function.FunctionData) step.Step {
	return updateFunc{
		fn:       fn,
		name:     name,
		funcData: data,
		log:      log,
	}
}
