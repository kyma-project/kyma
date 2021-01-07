package step

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SerialRunner struct {
	steps       []Step
	lastStepIdx int
	name        string
	log         *logrus.Entry
}

func NewSerialTestRunner(log *logrus.Entry, name string, steps ...Step) *SerialRunner {
	return &SerialRunner{log: log, steps: steps, name: name}
}

func (s SerialRunner) Name() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s, Steps: ", s.name))
	for i, v := range s.steps {
		builder.WriteString(fmt.Sprintf("%d:%s", i, v.Name()))
		if len(s.steps) != i+1 {
			builder.WriteString(", ")
		}
	}
	builder.WriteString(".")
	return builder.String()
}

func (s *SerialRunner) Run() error {
	for i, serialStep := range s.steps {
		s.log.Infof("Running Step %d: %s", i, serialStep.Name())
		if err := serialStep.Run(); err != nil {
			s.log.Errorf("Error in %s, error: %s", serialStep.Name(), err.Error())
			if callBackErr := s.stepsOnError(i); callBackErr != nil {
				s.log.Errorf("while executing OnError on %s,, err: %s", serialStep.Name(), callBackErr.Error())
			}
			s.lastStepIdx = i
			return errors.Wrapf(err, "while executing step: %s", serialStep.Name())
		}
	}
	s.lastStepIdx = len(s.steps) - 1
	return nil
}

func (s SerialRunner) stepsOnError(stepIdx int) error {
	var errAll *multierror.Error

	for i := stepIdx; i >= 0; i-- {
		err := s.steps[i].OnError()
		if err != nil {
			errAll = multierror.Append(errAll, err)
		}
	}
	return errAll.ErrorOrNil()
}

func (s SerialRunner) Cleanup() error {
	for i := s.lastStepIdx; i >= 0; i-- {
		serialStep := s.steps[i]
		s.log.Infof("Cleanup Serial Step: %s", serialStep.Name())
		if err := serialStep.Cleanup(); err != nil {
			s.log.Errorf("while clean up step: %s, error: %s", serialStep.Name(), err.Error())
		}
	}
	return nil
}

func (s SerialRunner) OnError() error {
	return nil
}

var _ Step = &SerialRunner{}
