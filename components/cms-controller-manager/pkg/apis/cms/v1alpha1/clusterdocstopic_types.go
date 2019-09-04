package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterDocsTopicSpec defines the desired state of ClusterDocsTopic
type ClusterDocsTopicSpec struct {
	CommonDocsTopicSpec `json:",inline"`
}

// ClusterDocsTopicStatus defines the observed state of ClusterDocsTopic
type ClusterDocsTopicStatus struct {
	CommonDocsTopicStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// ClusterDocsTopic is the Schema for the clusterdocstopics API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClusterDocsTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterDocsTopicSpec   `json:"spec,omitempty"`
	Status ClusterDocsTopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterDocsTopicList contains a list of ClusterDocsTopic
type ClusterDocsTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDocsTopic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterDocsTopic{}, &ClusterDocsTopicList{})
}
