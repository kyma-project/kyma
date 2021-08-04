package process

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ Step = &DeletePublisherDeployment{}

// DeletePublisherDeployment struct implements the interface Step
type DeletePublisherDeployment struct {
	name    string
	process *Process
}

// NewDeletePublisherDeployment returns new instance of NewDeletePublisherDeployment struct
func NewDeletePublisherDeployment(p *Process) DeletePublisherDeployment {
	return DeletePublisherDeployment{
		name:    "Delete eventing publisher deployment",
		process: p,
	}
}

// ToString returns step name
func (s DeletePublisherDeployment) ToString() string {
	return s.name
}

// Do deletes the eventing-publisher deployment
func (s DeletePublisherDeployment) Do() error {
	// Get eventing-controller deployment object
	err := s.process.Clients.Deployment.Delete(s.process.KymaNamespace, s.process.PublisherName)
	// Ignore the error if its 404 error
	if err != nil && !k8serrors.IsNotFound(err){
		return err
	}

	return nil
}
