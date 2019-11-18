// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	internalclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
	internalinterfaces "github.com/kyma-project/kyma/components/event-bus/client/generated/informer/externalversions/internalinterfaces"
	v1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/applicationconnector/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// EventActivationInformer provides access to a shared informer and lister for
// EventActivations.
type EventActivationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.EventActivationLister
}

type eventActivationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewEventActivationInformer constructs a new informer for EventActivation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEventActivationInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredEventActivationInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredEventActivationInformer constructs a new informer for EventActivation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredEventActivationInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().EventActivations(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().EventActivations(namespace).Watch(options)
			},
		},
		&applicationconnectorv1alpha1.EventActivation{},
		resyncPeriod,
		indexers,
	)
}

func (f *eventActivationInformer) defaultInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredEventActivationInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *eventActivationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&applicationconnectorv1alpha1.EventActivation{}, f.defaultInformer)
}

func (f *eventActivationInformer) Lister() v1alpha1.EventActivationLister {
	return v1alpha1.NewEventActivationLister(f.Informer().GetIndexer())
}
