package fake

import (
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Commander struct {
	Client  dynamic.Interface
	Backend handlers.MessagingBackend
}

func (c *Commander) Init(mgr manager.Manager) error {
	return nil
}

func (c *Commander) Start() error {
	return nil
}

func (c *Commander) Stop() error {
	return nil
}

type Cleaner struct {
}

func (c *Cleaner) Clean(eventType string) (string, error) {
	// Cleaning is not needed in this test
	return eventType, nil
}
