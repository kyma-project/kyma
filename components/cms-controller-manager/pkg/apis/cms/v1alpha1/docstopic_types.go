package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DocsTopicSpec defines the desired state of DocsTopic
type DocsTopicSpec struct {
	CommonDocsTopicSpec `json:",inline"`
}

// DocsTopicStatus defines the observed state of DocsTopic
type DocsTopicStatus struct {
	CommonDocsTopicStatus `json:",inline"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DocsTopic is the Schema for the docstopic API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DocsTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DocsTopicSpec   `json:"spec,omitempty"`
	Status DocsTopicStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DocsTopicList contains a list of DocsTopic
type DocsTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DocsTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DocsTopic{}, &DocsTopicList{})
}
