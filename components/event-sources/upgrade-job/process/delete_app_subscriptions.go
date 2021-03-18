package process

import (
	"github.com/pkg/errors"
)

var _ Step = &DeleteAppSubscriptions{}

const (
	applicationNameKey = "application-name"
)

type DeleteAppSubscriptions struct {
	name    string
	process *Process
}

func NewDeleteAppSubscriptions(p *Process) DeleteAppSubscriptions {
	return DeleteAppSubscriptions{
		name:    "Delete application subscriptions",
		process: p,
	}
}

func (s DeleteAppSubscriptions) Do() error {
	for _, sub := range s.process.State.Subscriptions.Items {
		if sub.Labels[applicationNameKey] != "" {
			err := s.process.Clients.Subscription.Delete(sub.Namespace, sub.Name)
			if err != nil {
				return errors.Wrapf(err, "failed to delete app subscription %s/%s", sub.Namespace, sub.Name)
			}
		}
		s.process.Logger.Infof("Step: %s, deleted knative subscription %s/%s to brokers", s.ToString(), sub.Namespace, sub.Name)
	}
	return nil
}

func (s DeleteAppSubscriptions) ToString() string {
	return s.name
}
