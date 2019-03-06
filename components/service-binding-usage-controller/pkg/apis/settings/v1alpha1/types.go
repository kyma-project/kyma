package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodPreset is the Schema for the podpresets API
// +k8s:openapi-gen=true
type PodPreset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodPresetSpec   `json:"spec,omitempty"`
	Status PodPresetStatus `json:"status,omitempty"`
}

// PodPresetSpec defines the desired state of PodPreset
type PodPresetSpec struct {
	// Selector is a label query over a set of resources, in this case pods.
	// Required.
	Selector metav1.LabelSelector `json:"selector"`

	// Env defines the collection of EnvVar to inject into containers.
	// +optional
	Env []v1.EnvVar `json:"env,omitempty"`

	// EnvFrom defines the collection of EnvFromSource to inject into containers.
	// +optional
	EnvFrom []v1.EnvFromSource `json:"envFrom,omitempty"`

	// Volumes defines the collection of Volume to inject into the pod.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`

	// VolumeMounts defines the collection of VolumeMount to inject into containers.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
}

// PodPresetStatus defines the observed state of PodPreset
type PodPresetStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PodPresetList contains a list of PodPreset
type PodPresetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodPreset `json:"items"`
}
