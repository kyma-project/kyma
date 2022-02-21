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

package v1beta1

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

type Source struct {
	// +optional
	Artifact *ArtifactSource `json:"artifact,omitempty"`
	//TODO: we should check if it is possible
	// +optional
	CRD *CRDSourceRef `json:"crd,omitempty"`
	// +optional
	GitRepository *GitRepositorySource `json:"gitRepository,omitempty"`
	// +optional
	Inline *InlineSource `json:"inline,omitempty"`
}

type InlineSource struct {
	Source       string `json:"source"`
	Dependencies string `json:"dependencies"`
}

type GitRepositorySource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type CRDSourceRef struct {
	Api        string `json:"api"`
	Version    string `json:"version"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespaces string `json:"namespaces"`
	Path       string `json:"path"`
	BaseDir    string `json:"baseDir"`
	// +optional
	Credentials *Credentials `json:"credentials,omitempty"`
}

type ArtifactSource struct {
	//TODO: maybe it's worth do distinguish between local and external, as we know the internal URL
	URL     string `json:"url"`
	BaseDir string `json:"baseDir"`
	// +optional
	Credentials *Credentials `json:"credentials,omitempty"`
}

type ConfigMapRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

//TODO: To discuss auth options
type Credentials struct {
	// +optional
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	// +optional
	Oauth *Oauth `json:"oauth,omitempty"`
	// +optional
	JWTToken *v1.SecretKeySelector `json:"token,omitempty"`
	// +optional
	PersonalAccessToken *v1.SecretKeySelector `json:"personalAccessToken,omitempty"`
}

//TODO: should we enforce using secret or give the ability to use configmaps or plain values
type BasicAuth struct {
	User     v1.SecretKeySelector `json:"user"`
	Password v1.SecretKeySelector `json:"password"`
}

type Oauth struct {
	OauthURL string               `json:"oatuhURL"`
	User     v1.SecretKeySelector `json:"user"`
	Password v1.SecretKeySelector `json:"password"`
}

type Template struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations"`
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
	RuntimeProfile *string `json:"runtimeProfile,omitempty"`

	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty,omitempty"`

	// +optional
	BuildProfile *string `json:"buildProfile,omitempty"`

	// +optional
	BuildResources *v1.ResourceRequirements `json:"buildResources,omitempty"`

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
// +kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='ConfigurationReady')].status"
// +kubebuilder:printcolumn:name="Built",type="string",JSONPath=".status.conditions[?(@.type=='BuildReady')].status"
// +kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
// +kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".spec.runtime"
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=".spec.source"
// +kubebuilder:printcolumn:name="Version",type="integer",JSONPath=".metadata.generation"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Function is the Schema for the functions API
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}
