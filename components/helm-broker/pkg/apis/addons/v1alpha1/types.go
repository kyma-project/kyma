package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true

// AddonsConfiguration is the Schema for the addonsconfigurations API
// Important: Run "make generates" to regenerate files after modifying this struct
//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:categories=all,addons
type AddonsConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonsConfigurationSpec   `json:"spec,omitempty"`
	Status AddonsConfigurationStatus `json:"status,omitempty"`
}

// AddonsConfigurationSpec defines the desired state of AddonsConfiguration
type AddonsConfigurationSpec struct {
	CommonAddonsConfigurationSpec `json:",inline"`
}

// AddonsConfigurationStatus defines the observed state of AddonsConfiguration
type AddonsConfigurationStatus struct {
	CommonAddonsConfigurationStatus `json:",inline"`
}

// AddonsConfigurationList contains a list of AddonsConfiguration
// Important: Run "make generates" to regenerate files after modifying this struct
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AddonsConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddonsConfiguration `json:"items"`
}

// +kubebuilder:object:root=true

// ClusterAddonsConfiguration is the Schema for the addonsconfigurations API
// Important: Run "make generates" to regenerate files after modifying this struct
//
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:categories=all,addons
type ClusterAddonsConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterAddonsConfigurationSpec `json:"spec,omitempty"`
	// +optional
	Status ClusterAddonsConfigurationStatus `json:"status,omitempty"`
}

// ClusterAddonsConfigurationSpec defines the desired state of ClusterAddonsConfiguration
type ClusterAddonsConfigurationSpec struct {
	CommonAddonsConfigurationSpec `json:",inline"`
}

// ClusterAddonsConfigurationStatus defines the observed state of ClusterAddonsConfiguration
type ClusterAddonsConfigurationStatus struct {
	CommonAddonsConfigurationStatus `json:",inline"`
}

// ClusterAddonsConfigurationList contains a list of ClusterAddonsConfiguration
// Important: Run "make generates" to regenerate files after modifying this struct
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterAddonsConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAddonsConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&AddonsConfiguration{}, &AddonsConfigurationList{},
		&ClusterAddonsConfiguration{}, &ClusterAddonsConfigurationList{},
	)
}
