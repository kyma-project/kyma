// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	v1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=addons.kyma-project.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("addonsconfigurations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Addons().V1alpha1().AddonsConfigurations().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("clusteraddonsconfigurations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Addons().V1alpha1().ClusterAddonsConfigurations().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
