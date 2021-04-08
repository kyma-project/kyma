package process

import (
	"github.com/pkg/errors"
)

var _ Step = &DeleteEventSources{}

type DeleteEventSources struct {
	name    string
	process *Process
}

func NewDeleteEventSources(p *Process) DeleteEventSources {
	return DeleteEventSources{
		name:    "Delete Event Sources",
		process: p,
	}
}

func (s DeleteEventSources) Do() error {
	for _, eventSource := range s.process.State.EventSources.Items {
		err := s.process.Clients.HttpSource.Delete(eventSource.Namespace, eventSource.Name)
		if err != nil {
			return errors.Wrapf(err, "failed to delete EventSource %s/%s", eventSource.Namespace, eventSource.Name)
		}
		s.process.Logger.Infof("Step: %s, deleted event-source %s/%s", s.ToString(), eventSource.Namespace, eventSource.Name)
	}
	return nil
}

func (s DeleteEventSources) ToString() string {
	return s.name
}
