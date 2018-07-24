package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group   = "config.istio.io"
	Version = "v1alpha2"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{
		Group:   Group,
		Version: Version,
	}

	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("rule"), &Rule{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("ruleList"), &RuleList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("denier"), &Denier{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("denierList"), &DenierList{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("checknothing"), &Checknothing{})
	scheme.AddKnownTypeWithName(SchemeGroupVersion.WithKind("checknothingList"), &ChecknothingList{})
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
