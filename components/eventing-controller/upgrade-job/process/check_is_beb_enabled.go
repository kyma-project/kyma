package process

import (
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
		secretList, err := s.checkIfBebEnabled124()
		if err != nil {
			return err
		}
		if s.process.State.IsBebEnabled {
			return s.process.Clients.EventMesh.InitUsingSecret(&secretList.Items[0])
		}
	} else {
		s.process.Logger.WithContext().Infof("Skipping step: %s, because it is not a v1.24.x cluster", s.ToString())
	}
	return nil
}

// checkIfBebEnabled124 checks if BEB is enabled in v1.24.x and initialises BEB client
// It also sets s.process.State.IsBebEnabled flag
func (s CheckIsBebEnabled) checkIfBebEnabled124() (*corev1.SecretList, error) {
	// Get BEB configs from beb k8s secret
	secretLabel := backend.BEBBackendSecretLabelKey + "=" + backend.BEBBackendSecretLabelValue
	secretList, err := s.process.Clients.Secret.ListByMatchingLabels(corev1.NamespaceAll, secretLabel)
	if err != nil {
		return nil, err
	}
	if len(secretList.Items) == 0 {
		s.process.State.IsBebEnabled = false
		return secretList, nil
	}
	if len(secretList.Items) > 1 {
		return nil, errors.New("more than 1 BEB secrets found")
	}

	s.process.State.IsBebEnabled = true
	return secretList, nil
}
