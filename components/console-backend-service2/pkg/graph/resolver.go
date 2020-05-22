package graph

import (
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

//go:generate genny -in=model/k8s_types.genny -out=model/k8s_types_gen.go gen "Value=Namespace,Application,ApplicationMapping,Pod,BackendModule"
//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*k8s.ApplicationConnectorServices
	*k8s.CoreServices
	backendModules *resource.Service
}

func NewResolver(serviceFactory *resource.ServiceFactory) *Resolver {
	backendModules := serviceFactory.ForResource(v1alpha1.SchemeGroupVersion.WithResource("backendmodules"))

	return &Resolver{
		CoreServices:                 k8s.NewCoreServices(serviceFactory),
		backendModules:               backendModules,
		ApplicationConnectorServices: k8s.NewApplicationConnectorServices(serviceFactory),
	}
}
