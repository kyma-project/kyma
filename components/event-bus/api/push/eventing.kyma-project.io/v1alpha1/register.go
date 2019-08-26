package v1alpha1

import (
	eventingkymaio "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: eventingkymaio.GroupName, Version: "v1alpha1"}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder instance
	SchemeBuilder runtime.SchemeBuilder
	// localSchemeBuilder instance
	localSchemeBuilder = &SchemeBuilder
	// AddToScheme instance
	AddToScheme = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Adds the list of known types to apis.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Subscription{},
		&SubscriptionList{})

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
