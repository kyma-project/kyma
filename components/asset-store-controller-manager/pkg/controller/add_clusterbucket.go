package controller

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/controller/clusterbucket"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clusterbucket.Add)
}
