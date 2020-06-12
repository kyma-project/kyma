package resource

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"time"
)

type ServiceFactory struct {
	Client          dynamic.Interface
	InformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func NewServiceFactoryForConfig(config *rest.Config, informerResyncPeriod time.Duration) (*ServiceFactory, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, informerResyncPeriod)
	return NewServiceFactory(dynamicClient, informerFactory), nil
}

func NewServiceFactory(client dynamic.Interface, informerFactory dynamicinformer.DynamicSharedInformerFactory) *ServiceFactory {
	return &ServiceFactory{
		Client:          client,
		InformerFactory: informerFactory,
	}
}

func (f *ServiceFactory) ForResource(gvr schema.GroupVersionResource) *Service {
	notifier := NewNotifier()
	informer := f.InformerFactory.ForResource(gvr).Informer()
	informer.AddEventHandler(notifier)
	return &Service{
		ServiceBase: &enabledServiceBase{
			Client:   f.Client.Resource(gvr),
			Informer: informer,
			notifier: notifier,
			gvr:      gvr,
		},
	}
}
