package subscriptionmanager

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

type Params map[string]interface{}

// Manager defines the interface that subscription managers for different messaging backends should implement.
type Manager interface {
	// Init inizializes the subscription manager and passes the controller manager to use.
	Init(mgr manager.Manager) error

	// Start runs the initialized subscription manager instance.
	Start(defaultSubsConfig env.DefaultSubscriptionConfig, params Params) error

	// Stop tells the subscription manager instance to shutdown and clean-up.
	Stop(runCleanup bool) error
}
