package teststep

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"strings"
)

type SerialSteps struct {
	steps []step.Step
	name  string
}

func (s SerialSteps) Name() string {
	builder := strings.Builder{}
	builder.WriteString(s.name)
	for i, v := range s.steps {
		builder.WriteString(fmt.Sprintf("%d: %s\n", i, v.Name()))
	}
	return builder.String()
}

func (s SerialSteps) Run() error {
	//TODO: Implementd
	panic("implement me")
}

func (s SerialSteps) Cleanup() error {
	panic("implement me")
}

func NewSerialSteps(name string, steps ...step.Step) SerialSteps {
	return SerialSteps{steps: steps, name: name}
}

var _ step.Step = SerialSteps{}
