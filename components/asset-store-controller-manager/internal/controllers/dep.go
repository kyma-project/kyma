package controllers

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/loader"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/store"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Container struct {
	Manager   ctrl.Manager
	Store     store.Store
	Loader    loader.Loader
	Validator assethook.Validator
	Mutator   assethook.Mutator
	Extractor assethook.MetadataExtractor
}
