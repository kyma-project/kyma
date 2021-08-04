package process

import eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

var _ Step = &CheckIsBebEnabled{}

// CheckIsBebEnabled struct implements the interface Step
type CheckIsBebEnabled struct {
	name    string
	process *Process
}

// NewCheckIsBebEnabled returns new instance of NewCheckIsBebEnabled struct
func NewCheckIsBebEnabled(p *Process) CheckIsBebEnabled {
	return CheckIsBebEnabled{
		name:    "Check if BEB enabled",
		process: p,
	}
}

// ToString returns step name
func (s CheckIsBebEnabled) ToString() string {
	return s.name
}

// Do checks if BEB is enabled in the Kyma Cluster and saves the result in process state
func (s CheckIsBebEnabled) Do() error {
	eventingbackendObject, err := s.process.Clients.EventingBackend.Get(s.process.KymaNamespace, "eventing-backend")
	if err != nil {
		return err
	}

	s.process.State.IsBebEnabled = false
	if eventingbackendObject.Status.Backend == eventingv1alpha1.BebBackendType {
		s.process.State.IsBebEnabled = true
	}

	return nil
}
