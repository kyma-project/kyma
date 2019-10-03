package resource

import (
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"time"
)

type ServiceFactory struct {
	Client          dynamic.Interface
	InformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func NewServiceFactory(config *rest.Config, informerResyncPeriod time.Duration) (*ServiceFactory, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, informerResyncPeriod)

	return &ServiceFactory{
		Client:          dynamicClient,
		InformerFactory: informerFactory,
	}, nil
}

func (f *ServiceFactory) ForResource(gvr schema.GroupVersionResource) *Service {
	return &Service{
		Client: f.Client.Resource(gvr),
		Informer: f.InformerFactory.ForResource(gvr).Informer(),
	}
}

type Service struct {
	Client    dynamic.NamespaceableResourceInterface
	Informer cache.SharedIndexInformer
}
