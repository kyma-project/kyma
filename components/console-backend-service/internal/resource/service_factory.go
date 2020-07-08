package resource

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
)

type GenericServiceFactory struct {
	Client          dynamic.Interface
	InformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func NewGenericServiceFactoryForConfig(config *rest.Config, informerResyncPeriod time.Duration) (*GenericServiceFactory, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, informerResyncPeriod)
	return NewGenericServiceFactory(dynamicClient, informerFactory), nil
}

func NewGenericServiceFactory(client dynamic.Interface, informerFactory dynamicinformer.DynamicSharedInformerFactory) *GenericServiceFactory {
	return &GenericServiceFactory{
		Client:          client,
		InformerFactory: informerFactory,
	}
}

func (f *GenericServiceFactory) ForResource(gvr schema.GroupVersionResource) *GenericService {
	notifier := NewNotifier()
	informer := f.
		InformerFactory.
		ForResource(gvr).
		Informer()
	informer.AddEventHandler(notifier)
	return &GenericService{
		ServiceBase: &enabledServiceBase{
			Client:   f.Client.Resource(gvr),
			Informer: informer,
			notifier: notifier,
			gvr:      gvr,
		},
	}
}
