package oauth

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
}

func New(factory *resource.GenericServiceFactory) *Resolver {
	module := resource.NewModule("apigateway", factory, resource.ServiceCreators{
		oAuth2ClientGroupVersionResource: NewService,
	})

	return &Resolver{
		Module: module,
	}
}

func (r *Resolver) Service() *resource.GenericService {
	return r.Module.Service(oAuth2ClientGroupVersionResource)
}
