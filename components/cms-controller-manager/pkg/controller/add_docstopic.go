package controller

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/controller/docstopic"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, docstopic.Add)
}
