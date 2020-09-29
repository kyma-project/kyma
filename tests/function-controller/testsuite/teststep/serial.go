package teststep

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type SerialSteps struct {
	steps []step.Step
	name  string
	log   *logrus.Entry
}

func (s SerialSteps) Name() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s, Steps: ", s.name))
	for i, v := range s.steps {
		// TODO: improve formatting
		builder.WriteString(fmt.Sprintf("%d:%s", i, v.Name()))
		if len(s.steps) != i+1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(".")
	return builder.String()
}

func (s SerialSteps) Run() error {
	for i, serialStep := range s.steps {
		s.log.Infof("Step %d: %s", i, serialStep.Name())
		if err := serialStep.Run(); err != nil {
			return errors.Wrapf(err, "while executing step: %s", serialStep.Name())
		}
	}
	return nil
}

func (s SerialSteps) Cleanup() error {
	for _, serialStep := range s.steps {
		s.log.Infof("Cleanup Serial Step: %s", serialStep.Name())
		if err := serialStep.Cleanup(); err != nil {
			return errors.Wrapf(err, "while clean up step: %s", serialStep.Name())
		}
	}
	return nil
}

func (s SerialSteps) OnError(cause error) error {
	for _, testStep := range s.steps {
		err := testStep.OnError(cause)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("while fetching logs from serial steps: %s", s.name))
		}
	}
	return nil
}

func NewSerialSteps(log *logrus.Entry, name string, steps ...step.Step) SerialSteps {
	return SerialSteps{log: log, steps: steps, name: name}
}

var _ step.Step = SerialSteps{}
