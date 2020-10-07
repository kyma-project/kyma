package k8sNew

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
}

func New(factory *resource.GenericServiceFactory) *Resolver {
	module := resource.NewModule("k8sNew", factory, resource.ServiceCreators{
		limitRangesGroupVersionResource: NewLimitRangesService,
	})
	// informerFactory := informers.NewSharedInformerFactory(clientset, informerResyncPeriod)

	return &Resolver{
		Module: module,
		// informerFactory: informerFactory,
	}
}

func (r *Resolver) LimitRangesService() *resource.GenericService {
	return r.Module.Service(limitRangesGroupVersionResource)
}

// func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
// 	r.informerFactory.Start(stopCh)
// 	r.informerFactory.WaitForCacheSync(stopCh)
// }
