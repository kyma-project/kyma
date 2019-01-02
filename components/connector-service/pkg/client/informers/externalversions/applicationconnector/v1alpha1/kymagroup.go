// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"
	versioned "github.com/kyma-project/kyma/components/connector-service/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kyma-project/kyma/components/connector-service/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kyma-project/kyma/components/connector-service/pkg/client/listers/applicationconnector/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// KymaGroupInformer provides access to a shared informer and lister for
// KymaGroups.
type KymaGroupInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.KymaGroupLister
}

type kymaGroupInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewKymaGroupInformer constructs a new informer for KymaGroup type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewKymaGroupInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredKymaGroupInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredKymaGroupInformer constructs a new informer for KymaGroup type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredKymaGroupInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().KymaGroups().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().KymaGroups().Watch(options)
			},
		},
		&applicationconnectorv1alpha1.KymaGroup{},
		resyncPeriod,
		indexers,
	)
}

func (f *kymaGroupInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredKymaGroupInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *kymaGroupInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&applicationconnectorv1alpha1.KymaGroup{}, f.defaultInformer)
}

func (f *kymaGroupInformer) Lister() v1alpha1.KymaGroupLister {
	return v1alpha1.NewKymaGroupLister(f.Informer().GetIndexer())
}
