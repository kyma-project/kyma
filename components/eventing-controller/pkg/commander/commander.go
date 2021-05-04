package commander

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Commander defines the interface of different implementations
type Commander interface {
	// Start runs the initialized commander instance using the given manager.
	Start(mgr manager.Manager) error

	// Stop tells the commander instance to shutdown and clean-up.
	Stop() error
}
