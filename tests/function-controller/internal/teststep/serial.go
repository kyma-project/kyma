package teststep

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/pkg/errors"
	"strings"
)

type SerialSteps struct {
	steps []step.Step
	name  string
	log   shared.Logger
}

func (s SerialSteps) Name() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s, Steps: ", s.name))
	for i, v := range s.steps {
		//TODO: improve formatting
		builder.WriteString(fmt.Sprintf("%d:%s", i, v.Name()))
		if len(s.steps) != i+1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(".")
	return builder.String()
}

func (s SerialSteps) Run() error {
	for _, serialStep := range s.steps {
		s.log.Logf(serialStep.Name())
		if err := serialStep.Run(); err != nil {
			return errors.Wrapf(err, "while excuting step: %s", serialStep.Name())
		}
	}
	return nil
}

func (s SerialSteps) Cleanup() error {
	for _, serialStep := range s.steps {
		s.log.Logf("Running Serial Step: %s", serialStep.Name())
		if err := serialStep.Cleanup(); err != nil {
			return errors.Wrapf(err, "while clean up step: %s", serialStep.Name())
		}
	}
	return nil
}

func NewSerialSteps(log shared.Logger, name string, steps ...step.Step) SerialSteps {
	return SerialSteps{log: log, steps: steps, name: name}
}

var _ step.Step = SerialSteps{}
