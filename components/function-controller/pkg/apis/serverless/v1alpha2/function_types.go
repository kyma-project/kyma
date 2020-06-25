package v1alpha2

import (
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceType string

const (
	Raw SourceType = "raw"
	Git SourceType = "git"
)

type Repository struct {
	BaseDir    string `json:"baseDir"`
	Dockerfile string `json:"dockerfile"`
	Commit     string `json:"commit"`
	Branch     string `json:"branch"`
}

/// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	v1alpha1.FunctionSpec `json:",inline"`

	// +kubebuilder:default:="raw"
	SourceType SourceType `json:"type"`

	Repository Repository `json:"repository,inline,omitempty"`
}

// Function is the Schema for the functions API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='ConfigurationReady')].status"
// +kubebuilder:printcolumn:name="Built",type="string",JSONPath=".status.conditions[?(@.type=='BuildReady')].status"
// +kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
// +kubebuilder:printcolumn:name="Version",type="integer",JSONPath=".metadata.generation"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FunctionSpec            `json:"spec,omitempty"`
	Status            v1alpha1.FunctionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}
