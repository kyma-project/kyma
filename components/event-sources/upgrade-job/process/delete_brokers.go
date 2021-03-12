package process

import (
	"github.com/pkg/errors"
)

var _ Step = &DeleteBrokers{}

type DeleteBrokers struct {
	name    string
	process *Process
}

func NewDeleteBrokers(p *Process) DeleteBrokers {
	return DeleteBrokers{
		name:    "Delete Brokers",
		process: p,
	}
}

func (s DeleteBrokers) Do() error {
	for _, broker := range s.process.State.Brokers.Items {
		err := s.process.Clients.Broker.Delete(broker.Namespace, broker.Name)
		if err != nil {
			return errors.Wrapf(err, "failed to delete broker %s/%s", broker.Namespace, broker.Name)
		}
		s.process.Logger.Infof("step: %s, deleted broker %s/%s", s.ToString(), broker.Namespace, broker.Name)
	}
	return nil
}

func (s DeleteBrokers) ToString() string {
	return s.name
}
