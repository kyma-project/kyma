// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	applicationconnector_v1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	versioned "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ApplicationMappingInformer provides access to a shared informer and lister for
// ApplicationMappings.
type ApplicationMappingInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ApplicationMappingLister
}

type applicationMappingInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewApplicationMappingInformer constructs a new informer for ApplicationMapping type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewApplicationMappingInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredApplicationMappingInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredApplicationMappingInformer constructs a new informer for ApplicationMapping type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredApplicationMappingInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().ApplicationMappings(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().ApplicationMappings(namespace).Watch(options)
			},
		},
		&applicationconnector_v1alpha1.ApplicationMapping{},
		resyncPeriod,
		indexers,
	)
}

func (f *applicationMappingInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredApplicationMappingInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *applicationMappingInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&applicationconnector_v1alpha1.ApplicationMapping{}, f.defaultInformer)
}

func (f *applicationMappingInformer) Lister() v1alpha1.ApplicationMappingLister {
	return v1alpha1.NewApplicationMappingLister(f.Informer().GetIndexer())
}
