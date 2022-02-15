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
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Runtime string

const (
	Python39 = "python39"
	NodeJs14 = "nodejs14"
)

type Source struct {
	CRD    *Crd    `json:"crd"`
	Http   *Http   `json:"http"`
	Inline *Inline `json:"inline"`
}

type Inline struct {
	Source       string `json:"source"`
	Dependencies string `json:"dependencies"`
}

type Crd struct {
	Api     string `json:"api"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Path    string `json:"path"`
}

type Http struct {
	URL     string `json:"url"`
	BaseDir string `json:"baseDir"`
	// +optional
	Credentials *Credentials `json:"credentials"`

	//+optional
	RetryInterval *time.Duration `json:"retryInterval"`
}

type AuthType string

const (
	OAuth AuthType = "oauth"
	Basic AuthType = "basic"
)

type Credentials struct {
	Type AuthType `json:"type"`
	//valueFrom:
	//secretKeyRef:
	//name: {{ template "docker-registry.fullname" . }}-secret
	//key: azureAccountName
	//*v1.SecretEnvSource
	OAuthURL string                `json:"OAuthURL"`
	User     *v1.SecretKeySelector `json:"user"`
	Password *v1.SecretKeySelector `json:"password"`
	Token    *v1.Secret            `json:"token"`
	//TODO: Cert authentication?
	//ValueFrom *v1.SecretKeySelector `json:"valueFrom"`
}

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Function. Edit function_types.go to remove/update
	Runtime `json:"runtime"`
	Source  `json:"source"`

	// Env defines an array of key value pairs need to be used as env variable for a function
	Env []v1.EnvVar `json:"env,omitempty"`

	Profile string `json:"profile"`

	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	BuildResources v1.ResourceRequirements `json:"buildResources,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
