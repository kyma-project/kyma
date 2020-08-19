package eventing

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/name"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
	generateName func() string
}

func New(factory *resource.GenericServiceFactory) *Resolver {
	module := resource.NewModule("eventing", factory, resource.ServiceCreators{
		triggersGroupVersionResource: NewService,
	})

	return &Resolver{
		Module:       module,
		generateName: name.Generate,
	}
}

func (r *Resolver) Service() *resource.GenericService {
	return r.Module.Service(triggersGroupVersionResource)
}
