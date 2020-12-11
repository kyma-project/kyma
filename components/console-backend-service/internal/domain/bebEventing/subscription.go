package bebEventing

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/client-go/kubernetes"
)

type Resolver struct {
	*resource.Module
	client *kubernetes.Clientset
}

func New(factory *resource.GenericServiceFactory, client *kubernetes.Clientset) *Resolver {
	module := resource.NewModule("eventing", factory, resource.ServiceCreators{
		subscriptionsGroupVersionResource: NewService,
		secretsGroupVersionResource:       NewSecretsService,
	})

	return &Resolver{
		Module: module,
		client: client,
	}
}

func (r *Resolver) Service() *resource.GenericService {
	return r.Module.Service(subscriptionsGroupVersionResource)
}

func (r *Resolver) SecretsService() *resource.GenericService {
	return r.Module.Service(secretsGroupVersionResource)
}
