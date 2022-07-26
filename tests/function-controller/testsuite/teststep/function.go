package teststep

import (
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type newFunction struct {
	name string
	spec serverlessv1alpha2.FunctionSpec
	fn   *function.Function
	log  *logrus.Entry
}

func CreateFunction(log *logrus.Entry, fn *function.Function, name string, spec serverlessv1alpha2.FunctionSpec) step.Step {
	return newFunction{
		fn:   fn,
		name: name,
		spec: spec,
		log:  log.WithField(step.LogStepKey, name),
	}
}

func (f newFunction) Name() string {
	return f.name
}

func (f newFunction) Run() error {
	if err := f.fn.Create(f.spec); err != nil {
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

type updateFunc struct {
	name     string
	funcData *function.FunctionData
	fn       *function.Function
	spec     serverlessv1alpha2.FunctionSpec
	log      *logrus.Entry
}

func UpdateFunction(log *logrus.Entry, fn *function.Function, name string, spec serverlessv1alpha2.FunctionSpec) step.Step {
	return updateFunc{
		fn:   fn,
		spec: spec,
		name: name,
		log:  log.WithField(step.LogStepKey, name),
	}
}

func (u updateFunc) Name() string {
	return u.name
}

func (u updateFunc) Run() error {
	if err := u.fn.Update(u.spec); err != nil {
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

type newFunctionv1alpha1 struct {
	name string
	spec serverlessv1alpha1.FunctionSpec
	fn   *v1alpha1.Function
	log  *logrus.Entry
}

func CreateFunctionV1Alpha1(log *logrus.Entry, fn *v1alpha1.Function, name string, spec serverlessv1alpha1.FunctionSpec) step.Step {
	return newFunctionv1alpha1{
		fn:   fn,
		name: name,
		spec: spec,
		log:  log.WithField(step.LogStepKey, name),
	}
}

func (f newFunctionv1alpha1) Name() string {
	return f.name
}

func (f newFunctionv1alpha1) Run() error {
	if err := f.fn.Create(f.spec); err != nil {
		return errors.Wrapf(err, "while creating function: %s", f.name)
	}

	f.log.Infof("Function Created, Waiting for ready status")
	return errors.Wrapf(f.fn.WaitForStatusRunning(), "while waiting for function: %s, to be ready:", f.name)
}

func (f newFunctionv1alpha1) Cleanup() error {
	return errors.Wrapf(f.fn.Delete(), "while deleting function: %s", f.name)
}

func (f newFunctionv1alpha1) OnError() error {
	return f.fn.LogResource()
}

var _ step.Step = newFunction{}
