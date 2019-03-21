package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterDocsTopicSpec defines the desired state of ClusterDocsTopic
type ClusterDocsTopicSpec struct {
	CommonDocsTopicSpec `json:",inline"`
}

// ClusterDocsTopicStatus defines the observed state of ClusterDocsTopic
type ClusterDocsTopicStatus struct {
	CommonDocsTopicStatus `json:",inline"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterDocsTopic is the Schema for the clusterdocstopic API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClusterDocsTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterDocsTopicSpec   `json:"spec,omitempty"`
	Status ClusterDocsTopicStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterDocsTopicList contains a list of ClusterDocsTopic
type ClusterDocsTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDocsTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterDocsTopic{}, &ClusterDocsTopicList{})
}
