package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterAssetSpec defines the desired state of ClusterAsset
type ClusterAssetSpec struct {
	CommonAssetSpec `json:",inline"`
}

// ClusterAssetStatus defines the observed state of ClusterAsset
type ClusterAssetStatus struct {
	CommonAssetStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// ClusterAsset is the Schema for the clusterassets API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Base URL",type="string",JSONPath=".status.assetRef.baseUrl"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClusterAsset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterAssetSpec   `json:"spec,omitempty"`
	Status ClusterAssetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterAssetList contains a list of ClusterAsset
type ClusterAssetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAsset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterAsset{}, &ClusterAssetList{})
}
