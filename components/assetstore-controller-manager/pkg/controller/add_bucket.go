package controller

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/bucket"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, bucket.Add)
}
