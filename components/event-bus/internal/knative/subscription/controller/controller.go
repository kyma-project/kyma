package controller

import (
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, *opts.Options) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, opts *opts.Options) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, opts); err != nil {
			return err
		}
	}
	return nil
}
