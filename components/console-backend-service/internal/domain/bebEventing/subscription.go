package bebEventing

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/client-go/kubernetes"
)

type Resolver struct {
	*resource.Module
	client kubernetes.Interface
}

func New(factory *resource.GenericServiceFactory, client kubernetes.Interface) *Resolver {
	module := resource.NewModule("eventing", factory, resource.ServiceCreators{
		subscriptionsGroupVersionResource: NewService,
	})

	return &Resolver{
		Module: module,
		client: client,
	}
}

func (r *Resolver) Service() *resource.GenericService {
	return r.Module.Service(subscriptionsGroupVersionResource)
}
