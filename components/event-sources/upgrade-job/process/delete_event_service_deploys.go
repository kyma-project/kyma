package process

import (
	"github.com/pkg/errors"
)

var _ Step = &DeleteEventServiceDeployments{}

type DeleteEventServiceDeployments struct {
	name    string
	process *Process
}

func NewDeleteEventServiceDeployments(p *Process) DeleteEventServiceDeployments {
	return DeleteEventServiceDeployments{
		name:    "Delete EventService Deployments",
		process: p,
	}
}

func (s DeleteEventServiceDeployments) Do() error {
	for _, deploy := range s.process.State.EventServices.Items {
		err := s.process.Clients.Deployment.Delete(deploy.Namespace, deploy.Name)
		if err != nil {
			return errors.Wrapf(err, "failed to delete deployment %s/%s", deploy.Namespace, deploy.Name)
		}
		s.process.Logger.Infof("Step: %s, deleted deployment %s/%s", s.ToString(), deploy.Namespace, deploy.Name)
	}

	return nil
}

func (s DeleteEventServiceDeployments) ToString() string {
	return s.name
}
