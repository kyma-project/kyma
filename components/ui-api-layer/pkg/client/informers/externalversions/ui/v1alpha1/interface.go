// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "github.com/kyma-project/kyma/components/ui-api-layer/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// BackendModules returns a BackendModuleInformer.
	BackendModules() BackendModuleInformer
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

// BackendModules returns a BackendModuleInformer.
func (v *version) BackendModules() BackendModuleInformer {
	return &backendModuleInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}
