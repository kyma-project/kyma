package graph

import (
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*k8s.ApplicationConnectorServices
	*k8s.CoreServices
	*k8s.UiServices
}

func NewResolver(serviceFactory *resource.ServiceFactory) *Resolver {
	return &Resolver{
		CoreServices:                 k8s.NewCoreServices(serviceFactory),
		UiServices:                   k8s.NewUiServices(serviceFactory),
		ApplicationConnectorServices: k8s.NewApplicationConnectorServices(serviceFactory),
	}
}
