// Code generated by informer-gen. DO NOT EDIT.

package v1alpha2

import (
	time "time"

	istiov1alpha2 "github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	versioned "github.com/kyma-project/kyma/components/application-registry/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kyma-project/kyma/components/application-registry/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha2 "github.com/kyma-project/kyma/components/application-registry/pkg/client/listers/istio/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// HandlerInformer provides access to a shared informer and lister for
// Handlers.
type HandlerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha2.HandlerLister
}

type handlerInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewHandlerInformer constructs a new informer for Handler type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewHandlerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredHandlerInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredHandlerInformer constructs a new informer for Handler type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredHandlerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IstioV1alpha2().Handlers(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IstioV1alpha2().Handlers(namespace).Watch(options)
			},
		},
		&istiov1alpha2.Handler{},
		resyncPeriod,
		indexers,
	)
}

func (f *handlerInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredHandlerInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *handlerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&istiov1alpha2.Handler{}, f.defaultInformer)
}

func (f *handlerInformer) Lister() v1alpha2.HandlerLister {
	return v1alpha2.NewHandlerLister(f.Informer().GetIndexer())
}
