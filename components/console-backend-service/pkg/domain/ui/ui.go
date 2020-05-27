package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"time"

	"k8s.io/client-go/rest"
)

type Resolver struct {
	*backendModuleResolver
	*microFrontendResolver
	*clusterMicroFrontendResolver
	serviceFactory *resource.ServiceFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Resolver, error) {
	sf, err := resource.NewServiceFactoryForConfig(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		backendModuleResolver: newBackendModuleResolver(sf),
		microFrontendResolver: newMicroFrontendResolver(sf),
		clusterMicroFrontendResolver: newClusterMicroFrontendResolver(sf),
		serviceFactory:        sf,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.serviceFactory.InformerFactory.Start(stopCh)
	r.serviceFactory.InformerFactory.WaitForCacheSync(stopCh)
}
