// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	internalclientset "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset"
	internalinterfaces "github.com/kyma-project/kyma/components/event-sources/client/generated/informer/externalversions/internalinterfaces"
	v1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/lister/sources/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// HTTPSourceInformer provides access to a shared informer and lister for
// HTTPSources.
type HTTPSourceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.HTTPSourceLister
}

type hTTPSourceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewHTTPSourceInformer constructs a new informer for HTTPSource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewHTTPSourceInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredHTTPSourceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredHTTPSourceInformer constructs a new informer for HTTPSource type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredHTTPSourceInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SourcesV1alpha1().HTTPSources(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SourcesV1alpha1().HTTPSources(namespace).Watch(options)
			},
		},
		&sourcesv1alpha1.HTTPSource{},
		resyncPeriod,
		indexers,
	)
}

func (f *hTTPSourceInformer) defaultInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredHTTPSourceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *hTTPSourceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&sourcesv1alpha1.HTTPSource{}, f.defaultInformer)
}

func (f *hTTPSourceInformer) Lister() v1alpha1.HTTPSourceLister {
	return v1alpha1.NewHTTPSourceLister(f.Informer().GetIndexer())
}
