package ui

import (
	"time"

	mfClient "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned"
	mfInformer "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/client/informers/externalversions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Container struct {
	Resolver                     *Resolver
	BackendModuleInformer        cache.SharedIndexInformer
	MicrofrontendInformer        cache.SharedIndexInformer
	ClusterMicrofrontendInformer cache.SharedIndexInformer
}

type Resolver struct {
	*backendModuleResolver
	*microfrontendResolver
	*clusterMicrofrontendResolver

	informerFactory              externalversions.SharedInformerFactory
	microfrontendInformerFactory mfInformer.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Container, error) {
	clientset, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	microfrontendClientset, err := mfClient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	informerFactory := externalversions.NewSharedInformerFactory(clientset, informerResyncPeriod)
	backendModuleInformer := informerFactory.Ui().V1alpha1().BackendModules().Informer()
	backendModuleService := newBackendModuleService(backendModuleInformer)

	microfrontendInformerFactory := mfInformer.NewSharedInformerFactory(microfrontendClientset, informerResyncPeriod)
	microfrontendInformer := microfrontendInformerFactory.Ui().V1alpha1().MicroFrontends().Informer()
	clusterMicrofrontendInformer := microfrontendInformerFactory.Ui().V1alpha1().ClusterMicroFrontends().Informer()
	microfrontendService := newMicrofrontendService(microfrontendInformer)
	clusterMicrofrontendService := newClusterMicrofrontendService(clusterMicrofrontendInformer)

	return &Container{
		Resolver: &Resolver{
			backendModuleResolver:        newBackendModuleResolver(backendModuleService),
			informerFactory:              informerFactory,
			microfrontendResolver:        newMicrofrontendResolver(microfrontendService),
			clusterMicrofrontendResolver: newClusterMicrofrontendResolver(clusterMicrofrontendService),
			microfrontendInformerFactory: microfrontendInformerFactory,
		},
		BackendModuleInformer:        backendModuleInformer,
		MicrofrontendInformer:        microfrontendInformer,
		ClusterMicrofrontendInformer: clusterMicrofrontendInformer,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
	r.microfrontendInformerFactory.Start(stopCh)
	r.microfrontendInformerFactory.WaitForCacheSync(stopCh)
}
