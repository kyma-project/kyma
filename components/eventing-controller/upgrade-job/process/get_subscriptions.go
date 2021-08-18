package process

import (
	corev1 "k8s.io/api/core/v1"
)

var _ Step = &GetSubscriptions{}

// GetSubscriptions struct implements the interface Step
type GetSubscriptions struct {
	name    string
	process *Process
}

// NewGetSubscriptions returns new instance of NewGetSubscriptions struct
func NewGetSubscriptions(p *Process) GetSubscriptions {
	return GetSubscriptions{
		name:    "Get list of subscriptions",
		process: p,
	}
}

// ToString returns step name
func (s GetSubscriptions) ToString() string {
	return s.name
}

// Do fetches all Kyma subscriptions and saves it to process state
func (s GetSubscriptions) Do() error {
	namespace := corev1.NamespaceAll

	subscriptionList, err := s.process.Clients.Subscription.List(namespace)
	if err != nil {
		return err
	}

	s.process.State.Subscriptions = subscriptionList
	return nil
}
