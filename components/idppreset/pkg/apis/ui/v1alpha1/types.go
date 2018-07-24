package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IDPPreset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              IDPPresetSpec `json:"spec"`
}

func (rem *IDPPreset) GetObjectKind() schema.ObjectKind {
	return &IDPPreset{}
}

type IDPPresetSpec struct {
	Name    string `json:"name"`
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IDPPresetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []IDPPreset `json:"items"`
}
