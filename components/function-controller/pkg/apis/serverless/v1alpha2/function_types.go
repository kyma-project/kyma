/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Runtime string

const (
	Python39 = "python39"
	NodeJs14 = "nodejs14"
)

type FunctionType int

const (
	FunctionTypeInline FunctionType = iota + 1
	FunctionTypeGit
)

var (
	ZeroGitRepository = GitRepositorySource{}
	ZeroInline        = InlineSource{}
)

type Source struct {
	// +optional
	GitRepository GitRepositorySource `json:"gitRepository,omitempty"`
	// +optional
	Inline InlineSource `json:"inline,omitempty"`
}

type InlineSource struct {
	Source       string `json:"source"`
	Dependencies string `json:"dependencies"`
}

type GitRepositorySource struct {
	// +kubebuilder:validation:Required

	// URL is the address of GIT repository
	URL string `json:"url"`

	// Auth is the optional definition of authentication that should be used for repository operations
	// +optional
	Auth *RepositoryAuth `json:"auth,omitempty"`
}

// RepositoryAuth defines authentication method used for repository operations
type RepositoryAuth struct {
	// Type is the type of authentication
	Type RepositoryAuthType `json:"type"`

	// +kubebuilder:validation:Required

	// SecretName is the name of Kubernetes Secret containing credentials used for authentication
	SecretName string `json:"secretName"`
}

// RepositoryAuthType is the enum of available authentication types
// +kubebuilder:validation:Enum=basic;key
type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey                    = "key"
)

type ConfigMapRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Template struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations"`
}

type ResourceRequirements struct {
	// +optional
	Profile *string `json:"profile,omitempty"`
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ResourceConfiguration struct {
	// +optional
	Build ResourceRequirements `json:"build"`
	// +optional
	Function ResourceRequirements `json:"function"`
}

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Function. Edit function_types.go to remove/update
	Runtime string `json:"runtime"`

	//+optional
	CustomRuntimeConfiguration *ConfigMapRef `json:"customRuntimeConfiguration,omitempty"`

	Source Source `json:"source"`

	// Env defines an array of key value pairs need to be used as env variable for a function
	Env []v1.EnvVar `json:"env,omitempty"`

	// +optional
	ResourceConfiguration ResourceConfiguration `json:"resourceConfiguration,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// +optional
	Template *Template `json:"template,omitempty"`
}

//TODO: Status related things needs to be developed.
type ConditionType string

const (
	ConditionRunning            ConditionType = "Running"
	ConditionConfigurationReady ConditionType = "ConfigurationReady"
	ConditionBuildReady         ConditionType = "BuildReady"
)

type ConditionReason string

type Condition struct {
	Type               ConditionType      `json:"type,omitempty"`
	Status             v1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason    `json:"reason,omitempty"`
	Message            string             `json:"message,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Runtime string `json:"runtime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='ConfigurationReady')].status"
//+kubebuilder:printcolumn:name="Built",type="string",JSONPath=".status.conditions[?(@.type=='BuildReady')].status"
//+kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
//+kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".spec.runtime"
//+kubebuilder:printcolumn:name="Source",type="string",JSONPath=".spec.source"
//+kubebuilder:printcolumn:name="Version",type="integer",JSONPath=".metadata.generation"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
}

func (f *Function) TypeOf(t FunctionType) bool {
	switch t {

	case FunctionTypeInline:
		return f.Spec.Source.Inline != ZeroInline

	case FunctionTypeGit:
		return f.Spec.Source.GitRepository != ZeroGitRepository

	default:
		return false
	}
}

//+kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&Function{},
		&FunctionList{},
	)
}
