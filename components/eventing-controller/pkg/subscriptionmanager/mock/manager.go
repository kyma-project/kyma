package mock

import (
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
)

type Manager struct {
	Client  dynamic.Interface
	Backend handlers.MessagingBackend
}

func (c *Manager) Init(mgr manager.Manager) error {
	return nil
}

func (c *Manager) Start(_ env.DefaultSubscriptionConfig, _ subscriptionmanager.Params) error {
	return nil
}

func (c *Manager) Stop(_ bool) error {
	return nil
}

type Cleaner struct {
}

func (c *Cleaner) Clean(eventType string) (string, error) {
	// Cleaning is not needed in this test
	return eventType, nil
}
