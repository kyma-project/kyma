package process

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ Step = &CheckClusterVersion{}

// CheckClusterVersion struct implements the interface Step
type CheckClusterVersion struct {
	name    string
	process *Process
}

// NewCheckClusterVersion returns new instance of NewCheckClusterVersion struct
func NewCheckClusterVersion(p *Process) CheckClusterVersion {
	return CheckClusterVersion{
		name:    "Check if the Kyma cluster is 1.24 or not",
		process: p,
	}
}

// ToString returns step name
func (s CheckClusterVersion) ToString() string {
	return s.name
}

// Do checks if BEB is enabled in the Kyma Cluster and saves the result in process state
// It also initializes the BEB client
func (s CheckClusterVersion) Do() error {
	is124, err := s.isClusterVersion124()
	s.process.State.Is124Cluster = is124
	return err
}

// isClusterVersion124 If it returns false, it is a 1.23 cluster
func (s CheckClusterVersion) isClusterVersion124() (bool, error) {
	eventingBackend, err := s.process.Clients.EventingBackend.Get(s.process.KymaNamespace, "eventing-backend")
	if err == nil {
		// eventing-backend instance found, meaning its a v1.24.x cluster
		s.process.Logger.WithContext().Info("found eventing backend instance %q", eventingBackend.Name)
		return true, nil
	}

	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}
