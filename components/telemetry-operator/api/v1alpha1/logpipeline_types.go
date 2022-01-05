/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LogPipelineSpec defines the desired state of LogPipeline.
type LogPipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Parsers          []Parser          `json:"parsers,omitempty"`
	MultiLineParsers []MultiLineParser `json:"multilineParsers,omitempty"`
	Filters          []Filter          `json:"filters,omitempty"`
	Outputs          []Output          `json:"outputs,omitempty"`
	Files            []FileMount       `json:"files,omitempty"`
	SecretRefs       []SecretReference `json:"secretRefs,omitempty"`
}

// Parser describes a Fluent Bit parser configuration section.
type Parser struct {
	Content string `json:"content,omitempty"`
}

// MultiLineParser describes a Fluent Bit multiline parser configuration section.
type MultiLineParser struct {
	Content string `json:"content,omitempty"`
}

// Filter describes a Fluent Bit filter configuration section.
type Filter struct {
	Content string `json:"content,omitempty"`
}

// Output describes a Fluent Bit output configuration section.
type Output struct {
	Content string `json:"content,omitempty"`
}

// FileMount provides file content to be consumed by a LogPipeline configuration.
type FileMount struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

// SecretReference is a pointer to a Kubernetes secret that should be provided as environment to Fluent Bit.
type SecretReference struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// LogPipelineStatus defines the observed state of LogPipeline.
type LogPipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// LogPipeline is the Schema for the logpipelines API
type LogPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogPipelineSpec   `json:"spec,omitempty"`
	Status LogPipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LogPipelineList contains a list of LogPipeline
type LogPipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogPipeline `json:"items"`
}

//nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&LogPipeline{}, &LogPipelineList{})
}
