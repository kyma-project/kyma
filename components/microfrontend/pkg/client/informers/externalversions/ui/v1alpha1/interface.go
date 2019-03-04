// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "github.com/kyma-project/kyma/components/microfrontend/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// ClusterMicroFrontends returns a ClusterMicroFrontendInformer.
	ClusterMicroFrontends() ClusterMicroFrontendInformer
	// MicroFrontends returns a MicroFrontendInformer.
	MicroFrontends() MicroFrontendInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// ClusterMicroFrontends returns a ClusterMicroFrontendInformer.
func (v *version) ClusterMicroFrontends() ClusterMicroFrontendInformer {
	return &clusterMicroFrontendInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// MicroFrontends returns a MicroFrontendInformer.
func (v *version) MicroFrontends() MicroFrontendInformer {
	return &microFrontendInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
