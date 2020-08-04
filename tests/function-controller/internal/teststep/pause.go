package teststep

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"time"
)

type Pause struct {
	time time.Duration
}

func (p Pause) Name() string {
	return "pause"
}

func (p Pause) Run() error {
	time.Sleep(p.time)
	return nil
}

func (p Pause) Cleanup() error {
	return nil
}

func NewPause(time time.Duration) Pause {
	return Pause{time: time}
}

var _ step.Step = Pause{}
