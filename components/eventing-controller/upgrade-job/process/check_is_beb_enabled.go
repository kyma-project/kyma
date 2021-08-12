package process

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/backend"
	corev1 "k8s.io/api/core/v1"
)

var _ Step = &CheckIsBebEnabled{}

// CheckIsBebEnabled struct implements the interface Step
type CheckIsBebEnabled struct {
	name    string
	process *Process
}

// NewCheckIsBebEnabled returns new instance of NewCheckIsBebEnabled struct
func NewCheckIsBebEnabled(p *Process) CheckIsBebEnabled {
	return CheckIsBebEnabled{
		name:    "Check if BEB enabled and Init BEB client",
		process: p,
	}
}

// ToString returns step name
func (s CheckIsBebEnabled) ToString() string {
	return s.name
}

// Do checks if BEB is enabled in the Kyma Cluster and saves the result in process state
// It also initializes the BEB client
func (s CheckIsBebEnabled) Do() error {
	// Set default to false
	s.process.State.IsBebEnabled = false

	if s.process.State.Is124Cluster {
		return s.CheckIfBebEnabled124()
	}

	s.process.Logger.WithContext().Info(fmt.Sprintf("Skipping step: %s, because it is not a v1.24.x cluster", s.ToString()))
	return nil
}

// CheckIfBebEnabled124 checks if BEB is enabled in v1.24.x and initialises BEB client
// It also sets s.process.State.IsBebEnabled flag
func (s CheckIsBebEnabled) CheckIfBebEnabled124() error {
	// Get BEB configs from beb k8s secret
	secretLabel := backend.BEBBackendSecretLabelKey + "=" + backend.BEBBackendSecretLabelValue
	secretList, err := s.process.Clients.Secret.ListByMatchingLabels(corev1.NamespaceAll, secretLabel)
	if err != nil {
		return err
	}
	if len(secretList.Items) == 0 {
		s.process.State.IsBebEnabled = false
		return nil
	}
	if len(secretList.Items) > 1 {
		return errors.New("more than 1 BEB secrets found")
	}

	s.process.State.IsBebEnabled = true
	return s.process.Clients.EventMesh.InitUsingSecret(&secretList.Items[0])
}
