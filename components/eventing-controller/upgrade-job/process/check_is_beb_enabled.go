package process

import (
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/backend"
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

	// Get eventing-backend CRD
	//eventingBackendCRD, err := s.process.Clients.EventingBackend.GetCRD()
	eventingBackend, err := s.process.Clients.EventingBackend.Get(s.process.KymaNamespace, "eventing-backend")
	if err == nil {
		// eventing-backend CRD found, meaning its a v1.24.x cluster
		//s.process.Logger.WithContext().Info(eventingBackendCRD.ObjectMeta.Name)
		s.process.Logger.WithContext().Info(eventingBackend.Name)

		s.process.State.Is124Cluster = true
		return s.CheckIfBebEnabled124()
	}

	if !k8serrors.IsNotFound(err) {
		return err
	}

	// eventing-backend CRD not found, meaning its a v1.23.x cluster
	s.process.State.Is124Cluster = false
	return s.CheckIfBebEnabled123()

	//if err == nil {
	//	// Check if backend set BEB
	//	if eventingBackendObject.Status.Backend == eventingv1alpha1.BebBackendType {
	//		s.process.State.IsBebEnabled = true
	//
	//		// Init BEB Client
	//		err = s.InitBebClientUsingBebSecret()
	//		if err != nil {
	//			return err
	//		}
	//		return nil
	//	}
	//	// else its a NATs cluster
	//	s.process.State.IsBebEnabled = false
	//	return nil
	//}
	//
	//// If there is any error other then 404 then return error
	//// else it means that its a older Kyma cluster (e.g. 1.23.x)
	//if !k8serrors.IsNotFound(err){
	//	return err
	//}
	//
	//// Logic to check isBEBEnabled for v1.23.x
	//// Get eventing secret
	//s.process.Logger.WithContext().Info("Checking for beb configs in secret: ", "eventing")
	//eventingSecret, err := s.process.Clients.Secret.Get(s.process.KymaNamespace, "eventing")
	//if err != nil {
	//	return err
	//}
	//
	//// If BEB config data in eventing secret is empty then it means
	//// that BEB is not enabled
	//bebNamespace, ok := eventingSecret.Data["beb-namespace"]
	//if !ok || string(bebNamespace) == "" {
	//	s.process.State.IsBebEnabled = false
	//	return nil
	//}
	//
	//s.process.State.IsBebEnabled = true
	//// Init BEB Client using eventing secret
	//err = s.InitBebClientUsingEventingSecret(eventingSecret)
	//if err != nil {
	//	return err
	//}
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

// CheckIfBebEnabled123 checks if BEB is enabled in v1.23.x and initialises BEB client
// It also sets s.process.State.IsBebEnabled
func (s CheckIsBebEnabled) CheckIfBebEnabled123() error {
	s.process.State.IsBebEnabled = false
	return nil
}
