package apigateway

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
}

func New(factory *resource.ServiceFactory) *Resolver {
	module := resource.NewModule("apigateway", factory, resource.ServiceCreators{
		apiRulesGroupVersionResource: NewService,
	})

	return &Resolver{
		Module: module,
	}
}

func (r *Resolver) Service() *resource.Service {
	return r.Module.Service(apiRulesGroupVersionResource)
}
