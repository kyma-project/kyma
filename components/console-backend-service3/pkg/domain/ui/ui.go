package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource"
	"time"

	"k8s.io/client-go/rest"
)

type Resolver struct {
	*backendModuleResolver
	*microFrontendResolver
	//*clusterMicroFrontendResolver

	serviceFactory *resource.ServiceFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Resolver, error) {
	sf, err := resource.NewServiceFactoryForConfig(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, err
	}

	//clientset, err := versioned.NewForConfig(restConfig)
	//if err != nil {
	//	return nil, err
	//}
	//microFrontendClientset, err := mfClient.NewForConfig(restConfig)
	//if err != nil {
	//	return nil, err
	//}

	//informerFactory := externalversions.NewSharedInformerFactory(clientset, informerResyncPeriod)
	//backendModuleInformer := informerFactory.Ui().V1alpha1().BackendModules().Informer()

	//microFrontendInformerFactory := mfInformer.NewSharedInformerFactory(microFrontendClientset, informerResyncPeriod)
	//microFrontendInformer := microFrontendInformerFactory.Ui().V1alpha1().MicroFrontends().Informer()
	//clusterMicroFrontendInformer := microFrontendInformerFactory.Ui().V1alpha1().ClusterMicroFrontends().Informer()
	//microFrontendService := newMicroFrontendService(microFrontendInformer)
	//clusterMicroFrontendService := newClusterMicroFrontendService(clusterMicroFrontendInformer)

	return &Resolver{
		backendModuleResolver: newBackendModuleResolver(sf),
		microFrontendResolver: newMicroFrontendResolver(sf),
		serviceFactory:        sf,
		//informerFactory:              informerFactory,
		//microFrontendResolver:        newMicroFrontendResolver(microFrontendService),
		//clusterMicroFrontendResolver: newClusterMicroFrontendResolver(clusterMicroFrontendService),
		//microFrontendInformerFactory: microFrontendInformerFactory,
		//BackendModuleInformer:        backendModuleInformer,
		//MicroFrontendInformer:        microFrontendInformer,
		//ClusterMicroFrontendInformer: clusterMicroFrontendInformer,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.serviceFactory.InformerFactory.Start(stopCh)
	r.serviceFactory.InformerFactory.WaitForCacheSync(stopCh)
}
