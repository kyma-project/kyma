package ui

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Container struct {
	Resolver                     *Resolver
	BackendModuleInformer        cache.SharedIndexInformer
	MicroFrontendInformer        cache.SharedIndexInformer
	ClusterMicroFrontendInformer cache.SharedIndexInformer
}

type Resolver struct {
	*backendModuleResolver
	*microFrontendResolver
	*clusterMicroFrontendResolver

	informerFactory              externalversions.SharedInformerFactory
	microFrontendInformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Container, error) {
	clientset, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	microFrontendClientset, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	informerFactory := externalversions.NewSharedInformerFactory(clientset, informerResyncPeriod)
	backendModuleInformer := informerFactory.Ui().V1alpha1().BackendModules().Informer()
	backendModuleService := newBackendModuleService(backendModuleInformer)

	microFrontendInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(microFrontendClientset, informerResyncPeriod)

	microFrontendInformer := microFrontendInformerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "microfrontends",
	}).Informer()

	clusterMicroFrontendInformer := microFrontendInformerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "clustermicrofrontends",
	}).Informer()

	microFrontendService := newMicroFrontendService(microFrontendInformer)
	clusterMicroFrontendService := newClusterMicroFrontendService(clusterMicroFrontendInformer)

	return &Container{
		Resolver: &Resolver{
			backendModuleResolver:        newBackendModuleResolver(backendModuleService),
			informerFactory:              informerFactory,
			microFrontendResolver:        newMicroFrontendResolver(microFrontendService),
			clusterMicroFrontendResolver: newClusterMicroFrontendResolver(clusterMicroFrontendService),
			microFrontendInformerFactory: microFrontendInformerFactory,
		},
		BackendModuleInformer:        backendModuleInformer,
		MicroFrontendInformer:        microFrontendInformer,
		ClusterMicroFrontendInformer: clusterMicroFrontendInformer,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
	r.microFrontendInformerFactory.Start(stopCh)
	r.microFrontendInformerFactory.WaitForCacheSync(stopCh)
}
