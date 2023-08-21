package function

import (
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

type newFunction struct {
	name string
	spec serverlessv1alpha2.FunctionSpec
	fn   *Function
	log  *logrus.Entry
}

func CreateFunction(log *logrus.Entry, fn *Function, name string, spec serverlessv1alpha2.FunctionSpec) executor.Step {
	return newFunction{
		fn:   fn,
		name: name,
		spec: spec,
		log:  log.WithField(executor.LogStepKey, name),
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

var _ executor.Step = newFunction{}

type updateFunc struct {
	name string
	fn   *Function
	spec serverlessv1alpha2.FunctionSpec
	log  *logrus.Entry
}

func UpdateFunction(log *logrus.Entry, fn *Function, name string, spec serverlessv1alpha2.FunctionSpec) executor.Step {
	return updateFunc{
		fn:   fn,
		spec: spec,
		name: name,
		log:  log.WithField(executor.LogStepKey, name),
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

var _ executor.Step = updateFunc{}
