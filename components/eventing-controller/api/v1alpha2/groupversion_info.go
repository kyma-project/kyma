// Package v1alpha2 contains API Schema definitions for the eventing v1alpha2 API group
// +kubebuilder:object:generate=true
// +groupName=eventing.kyma-project.io
package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "eventing.kyma-project.io", Version: "v1alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	//nolint:gochecknoglobals // required for tests
	// GroupKind is group kind to identify these objects.
	GroupKind = schema.GroupKind{Group: "eventing.kyma-project.io", Kind: "Subscription"}
)
