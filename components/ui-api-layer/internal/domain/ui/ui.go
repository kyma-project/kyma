package ui

import (
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/informers/externalversions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Container struct {
	Resolver              *Resolver
	BackendModuleInformer cache.SharedIndexInformer
}

type Resolver struct {
	*backendModuleResolver

	informerFactory externalversions.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Container, error) {
	clientset, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	informerFactory := externalversions.NewSharedInformerFactory(clientset, informerResyncPeriod)
	backendModuleInformer := informerFactory.Ui().V1alpha1().BackendModules().Informer()
	backendModuleService := newBackendModuleService(backendModuleInformer)

	return &Container{
		Resolver: &Resolver{
			backendModuleResolver: newBackendModuleResolver(backendModuleService),
			informerFactory:       informerFactory,
		},
		BackendModuleInformer: backendModuleInformer,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
