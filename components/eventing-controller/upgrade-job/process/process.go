package process

import (
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/common/logging/logger"
)

// Process contains the common resources for this upgrade-job steps
type Process struct {
	Steps              []Step
	ReleaseName        string
	Domain             string
	KymaNamespace      string
	ControllerName     string
	PublisherName      string
	PublisherNamespace string
	Clients            Clients
	State              State
	TimeoutPeriod      time.Duration
	Logger             *logger.Logger
}

// Execute runs the Do() method of each step
func (p Process) Execute() error {
	for _, s := range p.Steps {
		err := s.Do()
		if err != nil {
			return errors.Wrapf(err, "failed in step: %s", s.ToString())
		}
	}
	return nil
}

// AddSteps defines the execution sequence of steps for upgrade-job
func (p *Process) AddSteps() {
	p.Steps = []Step{
		NewScaleDownEventingController(p),
		NewDeletePublisherDeployment(p),
		NewGetSubscriptions(p),
		NewFilterSubscriptions(p),
		NewDeleteBebSubscriptions(p),
	}
}
