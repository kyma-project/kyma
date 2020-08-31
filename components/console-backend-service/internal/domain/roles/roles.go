package roles

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
}

func New(factory *resource.GenericServiceFactory) *Resolver {
	fmt.Println(roleGroupVersionResource)
	module := resource.NewModule("roles", factory, resource.ServiceCreators{
		roleGroupVersionResource: NewService,
	})

	return &Resolver{
		Module: module,
	}
}

func (r *Resolver) Service() *resource.GenericService {
	return r.Module.Service(roleGroupVersionResource)
}
