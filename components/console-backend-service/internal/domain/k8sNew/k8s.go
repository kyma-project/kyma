package k8sNew

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
}

func New(factory *resource.GenericServiceFactory) *Resolver {
	module := resource.NewModule("k8sNew", factory, resource.ServiceCreators{
		limitRangesGroupVersionResource:    NewLimitRangesService,
		resourceQuotasGroupVersionResource: NewResourceQuotasService,
	})

	return &Resolver{
		Module: module,
	}
}

func (r *Resolver) LimitRangesService() *resource.GenericService {
	return r.Module.Service(limitRangesGroupVersionResource)
}
func (r *Resolver) ResourceQuotasService() *resource.GenericService {
	return r.Module.Service(resourceQuotasGroupVersionResource)
}
