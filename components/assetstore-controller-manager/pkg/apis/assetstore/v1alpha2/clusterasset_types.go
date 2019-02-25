package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterAsset is the Schema for the clusterassets API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ClusterAsset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterAssetSpec   `json:"spec,omitempty"`
	Status ClusterAssetStatus `json:"status,omitempty"`
}

// ClusterAssetSpec defines the desired state of Cluster Asset
type ClusterAssetSpec struct {
	CommonAssetSpec `json:",inline"`
}

// ClusterAssetStatus defines the observed state of Cluster Asset
type ClusterAssetStatus struct {
	CommonAssetStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ClusterAssetList contains a list of ClusterAsset
type ClusterAssetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAsset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterAsset{}, &ClusterAssetList{})
}
