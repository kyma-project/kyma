package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kind represents Kubernetes Kind name
type Kind string

// TargetKindSpec defines the desired state of TargetKind
type TargetKindSpec struct {
	DisplayName string    `json:"displayName"`
	Resource    Resource `json:"resource"`
	LabelsPath  string    `json:"labelsPath"`
}

type Resource struct {
	Group   string `json:"group"`
	Kind    Kind `json:"kind"`
	Version string `json:"version"`
}

// TargetKindStatus defines the observed state of TargetKind
type TargetKindStatus struct {
	Registered bool `json:"registered"`
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
