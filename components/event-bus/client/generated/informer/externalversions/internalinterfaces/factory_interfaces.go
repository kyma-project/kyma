// Code generated by informer-gen. DO NOT EDIT.

package internalinterfaces

import (
	time "time"

	internalclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	cache "k8s.io/client-go/tools/cache"
)

// NewInformerFunc takes internalclientset.Interface and time.Duration to return a SharedIndexInformer.
type NewInformerFunc func(internalclientset.Interface, time.Duration) cache.SharedIndexInformer

// SharedInformerFactory a small interface to allow for adding an informer without an import cycle
type SharedInformerFactory interface {
	Start(stopCh <-chan struct{})
	InformerFor(obj runtime.Object, newFunc NewInformerFunc) cache.SharedIndexInformer
}

// TweakListOptionsFunc is a function that transforms a v1.ListOptions.
type TweakListOptionsFunc func(*v1.ListOptions)
