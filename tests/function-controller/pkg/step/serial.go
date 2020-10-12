package step

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type SerialRunner struct {
	steps []Step
	name  string
	log   *logrus.Entry
}

//TODO: Write test if steps are correclty executed and OnError also
func NewSerialTestRunner(log *logrus.Entry, name string, steps ...Step) SerialRunner {
	return SerialRunner{log: log, steps: steps, name: name}
}

func (s SerialRunner) Name() string {
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

func (s SerialRunner) Run() error {
	for i, serialStep := range s.steps {
		s.log.Infof("Step %d: %s", i, serialStep.Name())
		if err := serialStep.Run(); err != nil {
			if callBackErr := s.stepsOnError(err, i); callBackErr != nil {
				s.log.Errorf("while executing OnError on %s,, err: %s", serialStep.Name(), callBackErr.Error())
			}
			return errors.Wrapf(err, "while executing step: %s", serialStep.Name())
		}
	}
	return nil
}

func (s SerialRunner) stepsOnError(cause error, stepIdx int) error {
	for i := stepIdx; i >= 0; i-- {
		err := s.steps[i].OnError(cause)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s SerialRunner) Cleanup() error {
	for i := len(s.steps) - 1; i >= 0; i-- {
		serialStep := s.steps[i]
		s.log.Infof("Cleanup Serial Step: %s", serialStep.Name())
		if err := serialStep.Cleanup(); err != nil {
			return errors.Wrapf(err, "while clean up step: %s", serialStep.Name())
		}
	}
	return nil
}

func (s SerialRunner) OnError(cause error) error {
	return nil
}

var _ Step = SerialRunner{}
