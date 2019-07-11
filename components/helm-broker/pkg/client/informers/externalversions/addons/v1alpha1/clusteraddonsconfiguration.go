// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	versioned "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned"
	internalinterfaces "github.com/kyma-project/kyma/components/helm-broker/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/client/listers/addons/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterAddonsConfigurationInformer provides access to a shared informer and lister for
// ClusterAddonsConfigurations.
type ClusterAddonsConfigurationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ClusterAddonsConfigurationLister
}

type clusterAddonsConfigurationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewClusterAddonsConfigurationInformer constructs a new informer for ClusterAddonsConfiguration type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewClusterAddonsConfigurationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredClusterAddonsConfigurationInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredClusterAddonsConfigurationInformer constructs a new informer for ClusterAddonsConfiguration type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredClusterAddonsConfigurationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AddonsV1alpha1().ClusterAddonsConfigurations().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AddonsV1alpha1().ClusterAddonsConfigurations().Watch(options)
			},
		},
		&addonsv1alpha1.ClusterAddonsConfiguration{},
		resyncPeriod,
		indexers,
	)
}

func (f *clusterAddonsConfigurationInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterAddonsConfigurationInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *clusterAddonsConfigurationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&addonsv1alpha1.ClusterAddonsConfiguration{}, f.defaultInformer)
}

func (f *clusterAddonsConfigurationInformer) Lister() v1alpha1.ClusterAddonsConfigurationLister {
	return v1alpha1.NewClusterAddonsConfigurationLister(f.Informer().GetIndexer())
}
