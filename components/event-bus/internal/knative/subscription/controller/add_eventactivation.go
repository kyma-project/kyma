package controller

import (
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/eventactivation"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, eventactivation.Add)
}
