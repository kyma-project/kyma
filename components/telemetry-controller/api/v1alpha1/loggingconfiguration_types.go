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

// LoggingConfigurationSpec defines the desired state of LoggingConfiguration
type LoggingConfigurationSpec struct {
	Sections []Section `json:"sections,omitempty"`
}

// Section describes a Fluent Bit configuration section
type Section struct {
	Content     string            `json:"content,omitempty"`
	Environment []SecretReference `json:"environment,omitempty"`
	Files       []FileMount       `json:"files,omitempty"`
}

// FileMount provides file content to be consumed by a Section configuration
type FileMount struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

// SecretReference is a pointer to a Kubernetes secret that should be provided as environment to Fluent Bit
type SecretReference struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// LoggingConfigurationStatus defines the observed state of LoggingConfiguration
type LoggingConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LoggingConfiguration is the Schema for the loggingconfigurations API
type LoggingConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoggingConfigurationSpec   `json:"spec,omitempty"`
	Status LoggingConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LoggingConfigurationList contains a list of LoggingConfiguration
type LoggingConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoggingConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoggingConfiguration{}, &LoggingConfigurationList{})
}
