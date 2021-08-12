package process

import "errors"

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
	// First check if BEB client is initialized or not
	if !s.process.Clients.EventMesh.IsInitialised() {
		return errors.New("event mesh (BEB) client is not initialised")
	}

	// Traverse through the subscriptions and migrate
	subscriptionListItems := s.process.State.FilteredSubscriptions.Items

	for _, subscription := range subscriptionListItems {
		s.process.Logger.WithContext().Info("Deleting Event Mesh (BEB) Subscription: ", subscription.Name)
		_, err := s.process.Clients.EventMesh.Delete(subscription.Name)
		if err != nil {
			s.process.Logger.WithContext().Error(err)
			continue
		}
		// s.process.Logger.WithContext().Info(result.StatusCode, result.Message)
	}

	return nil
}
