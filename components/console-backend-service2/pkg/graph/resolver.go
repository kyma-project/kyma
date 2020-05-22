package graph

import (
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate genny -in=model/k8s_types.genny -out=model/k8s_types_gen.go gen "Value=Namespace,Application,ApplicationMapping,Pod"
//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	*k8s.ApplicationConnectorServices
	namespaces *resource.Service
	pods       *resource.Service
}

func NewResolver(serviceFactory *resource.ServiceFactory) *Resolver {
	namespaces := serviceFactory.ForResource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	})

	pods := serviceFactory.ForResource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	})

	return &Resolver{
		namespaces:                   namespaces,
		pods:                         pods,
		ApplicationConnectorServices: k8s.NewApplicationConnectorServices(serviceFactory),
	}
}
