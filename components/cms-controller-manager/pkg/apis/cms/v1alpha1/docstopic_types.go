package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DocsTopicSpec defines the desired state of DocsTopic
type DocsTopicSpec struct {
	CommonDocsTopicSpec `json:",inline"`
}

// DocsTopicStatus defines the observed state of DocsTopic
type DocsTopicStatus struct {
	CommonDocsTopicStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// DocsTopic is the Schema for the docstopics API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DocsTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DocsTopicSpec   `json:"spec,omitempty"`
	Status DocsTopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DocsTopicList contains a list of DocsTopic
type DocsTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DocsTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DocsTopic{}, &DocsTopicList{})
}
