package graph

import (
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*k8s.ApplicationConnectorServices
	namespaces          *resource.Service
}

func NewResolver(serviceFactory *resource.ServiceFactory) *Resolver {
	namespaces := serviceFactory.ForResource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	})

	return &Resolver{
		namespaces: namespaces,
		ApplicationConnectorServices: k8s.NewApplicationConnectorServices(serviceFactory),
	}
}
