package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TargetKindSpec defines the desired state of TargetKind
type TargetKindSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of TargetKind. Edit TargetKind_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// TargetKindStatus defines the observed state of TargetKind
type TargetKindStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// TargetKind is the Schema for the targetkinds API
type TargetKind struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetKindSpec   `json:"spec,omitempty"`
	Status TargetKindStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TargetKindList contains a list of TargetKind
type TargetKindList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetKind `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TargetKind{}, &TargetKindList{})
}
