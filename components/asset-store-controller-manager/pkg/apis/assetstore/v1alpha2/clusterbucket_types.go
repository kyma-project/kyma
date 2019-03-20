package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterBucket is the Schema for the clusterbuckets API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClusterBucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterBucketSpec   `json:"spec,omitempty"`
	Status ClusterBucketStatus `json:"status,omitempty"`
}

// ClusterBucketSpec defines the desired state of Bucket
type ClusterBucketSpec struct {
	CommonBucketSpec `json:",inline"`
}

// ClusterBucketStatus defines the observed state of Bucket
type ClusterBucketStatus struct {
	CommonBucketStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterBucketList contains a list of ClusterBucket
type ClusterBucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterBucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterBucket{}, &ClusterBucketList{})
}
