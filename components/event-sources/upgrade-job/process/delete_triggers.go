package process

import (
	"github.com/pkg/errors"
)

var _ Step = &DeleteTriggers{}

type DeleteTriggers struct {
	name    string
	process *Process
}

func NewDeleteTriggers(p *Process) DeleteTriggers {
	return DeleteTriggers{
		name:    "Delete Triggers",
		process: p,
	}
}

func (s DeleteTriggers) Do() error {
	for _, trigger := range s.process.State.Triggers.Items {

		err := s.process.Clients.Trigger.Delete(trigger.Namespace, trigger.Name)
		if err != nil {
			return errors.Wrapf(err, "failed to delete trigger %s/%s", trigger.Namespace, trigger.Name)
		}
		s.process.Logger.Infof("Step: %s, deleted trigger %s/%s", s.ToString(), trigger.Namespace, trigger.Name)
	}
	return nil
}

func (s DeleteTriggers) ToString() string {
	return s.name
}
