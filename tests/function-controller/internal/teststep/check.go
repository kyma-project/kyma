package teststep

import (
	"errors"
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"io/ioutil"
	"net/http"
)

type FunctionCheck struct {
	name        string
	url         string
	port        uint32
	expectedMsg string
	log         shared.Logger
}

func (f FunctionCheck) Name() string {
	return f.name
}

func (f FunctionCheck) Run() error {
	resp, err := http.Get(f.url)
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	out := string(bytes)

	f.log.Logf(`Got: "%s" response`, out)
	if f.expectedMsg != out {
		return fmt.Errorf(`expected: "%s", actual: "%s"`, f.expectedMsg, out)
	}
	return nil
}

func (f FunctionCheck) Cleanup() error {
	return nil
}

func NewCheck(log shared.Logger, name, url, expectedMsg string) *FunctionCheck {
	return &FunctionCheck{
		name:        name,
		url:         url,
		expectedMsg: expectedMsg,
		log:         log,
	}

}

var _ step.Step = FunctionCheck{}
var _ step.Step = DefaultedFunctionCheck{}

type DefaultedFunctionCheck struct {
	name string
	fn   *function.Function
}

func NewDefaultedFunctionCheck(fn *function.Function) step.Step {
	return &DefaultedFunctionCheck{
		name: "Check if function has set correctly default values",
		fn:   fn,
	}
}

func (e DefaultedFunctionCheck) Name() string {
	return e.name
}

func (e DefaultedFunctionCheck) Run() error {
	fn, err := e.fn.Get()
	if err != nil {
		return err
	}

	if fn == nil {
		return errors.New("function can't be nil")
	}

	spec := fn.Spec
	if spec.MinReplicas == nil {
		return errors.New("minReplicas equal nil")
	} else if spec.MaxReplicas == nil {
		return errors.New("maxReplicas equal nil")
	} else if spec.Resources.Requests.Memory().IsZero() || spec.Resources.Requests.Cpu().IsZero() {
		return errors.New("requests equal zero")
	} else if spec.Resources.Limits.Memory().IsZero() || spec.Resources.Limits.Cpu().IsZero() {
		return errors.New("limits equal zero")
	}
	return nil
}

func (e DefaultedFunctionCheck) Cleanup() error {
	return nil
}
