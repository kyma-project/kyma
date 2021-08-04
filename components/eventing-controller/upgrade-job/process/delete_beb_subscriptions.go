package process

import (
	"errors"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/backend"
	corev1 "k8s.io/api/core/v1"
)

var _ Step = &DeleteBebSubscriptions{}

// DeleteBebSubscriptions struct implements the interface Step
type DeleteBebSubscriptions struct {
	name    string
	process *Process
}

// NewDeleteBebSubscriptions returns new instance of NewDeleteBebSubscriptions struct
func NewDeleteBebSubscriptions(p *Process) DeleteBebSubscriptions {
	return DeleteBebSubscriptions{
		name:    "Delete BEB subscriptions from Event-Mesh",
		process: p,
	}
}

// ToString returns step name
func (s DeleteBebSubscriptions) ToString() string {
	return s.name
}

// Do deletes the subscriptions from Event Mesh (BEB)
func (s DeleteBebSubscriptions) Do() error {
	// First initialize the BEB client
	// Get BEB configs from beb k8s secret
	secretLabel := backend.BEBBackendSecretLabelKey + "=" + backend.BEBBackendSecretLabelValue
	secretList, err := s.process.Clients.Secret.ListByMatchingLabels(corev1.NamespaceAll, secretLabel)
	if err != nil {
		return err
	}
	if len(secretList.Items) == 0 {
		return errors.New("no BEB secrets found")
	}
	if len(secretList.Items) > 1 {
		return errors.New("more than 1 BEB secrets found")
	}

	// Initialize BEB client with this secret
	err = s.process.Clients.EventMesh.Init(&secretList.Items[0])
	if err != nil {
		return err
	}

	// Traverse through the subscriptions and migrate
	subscriptionListItems := s.process.State.FilteredSubscriptions.Items

	for _, subscription := range subscriptionListItems {
		s.process.Logger.WithContext().Info("Deleting: ", subscription.Name)
		result, err := s.process.Clients.EventMesh.Delete(subscription.Name)
		if err != nil {
			s.process.Logger.WithContext().Error(err)
			continue
		}

		s.process.Logger.WithContext().Info(result.StatusCode, result.Message)

		// @TODO: should we check for response
	}

	return nil
}
