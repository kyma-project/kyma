package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Asset is the Schema for the assets API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Base URL",type="string",JSONPath=".status.assetRef.baseUrl"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Asset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AssetSpec   `json:"spec,omitempty"`
	Status AssetStatus `json:"status,omitempty"`
}

// AssetSpec defines the desired state of Asset
type AssetSpec struct {
	CommonAssetSpec `json:",inline"`
}

// AssetStatus defines the observed state of Asset
type AssetStatus struct {
	CommonAssetStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AssetList contains a list of Asset
type AssetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Asset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Asset{}, &AssetList{})
}
