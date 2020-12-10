package bebEventing

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
}

func New(factory *resource.GenericServiceFactory) *Resolver {
	module := resource.NewModule("eventing", factory, resource.ServiceCreators{
		subscriptionsGroupVersionResource: NewService,
		secretsGroupVersionResource:       NewSecretsService,
	})

	return &Resolver{
		Module: module,
	}
}

func (r *Resolver) Service() *resource.GenericService {
	return r.Module.Service(subscriptionsGroupVersionResource)
}

func (r *Resolver) SecretsService() *resource.GenericService {
	return r.Module.Service(secretsGroupVersionResource)
}
