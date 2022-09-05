// Package v1alpha1 contains API Schema definitions for the eventing v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=eventing.kyma-project.io
package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/scheme"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription"
)

var (
	// GroupVersion is group version used to register these objects

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: subscription.GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
