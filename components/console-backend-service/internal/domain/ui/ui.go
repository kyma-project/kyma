package ui

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	v1 "github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
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

	informerFactory              dynamicinformer.DynamicSharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Container, error) {
	clientset, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(clientset, informerResyncPeriod)

	microFrontendInformer := informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "microfrontends",
	}).Informer()

	clusterMicroFrontendInformer := informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "clustermicrofrontends",
	}).Informer()

	backendModuleInformer := informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1.SchemeGroupVersion.Version,
		Group:    v1.SchemeGroupVersion.Group,
		Resource: "backendmodules",
	}).Informer()

	microFrontendService := newMicroFrontendService(microFrontendInformer)
	clusterMicroFrontendService := newClusterMicroFrontendService(clusterMicroFrontendInformer)
	backendModuleService := newBackendModuleService(backendModuleInformer)

	return &Container{
		Resolver: &Resolver{
			backendModuleResolver:        newBackendModuleResolver(backendModuleService),
			informerFactory:              informerFactory,
			microFrontendResolver:        newMicroFrontendResolver(microFrontendService),
			clusterMicroFrontendResolver: newClusterMicroFrontendResolver(clusterMicroFrontendService),
		},
		BackendModuleInformer:        backendModuleInformer,
		MicroFrontendInformer:        microFrontendInformer,
		ClusterMicroFrontendInformer: clusterMicroFrontendInformer,
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.informerFactory.Start(stopCh)
	r.informerFactory.WaitForCacheSync(stopCh)
}
