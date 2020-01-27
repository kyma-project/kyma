/*
Copyright 2019 The Kyma Authors.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// function defines the content of a function
	Function string `json:"function"`

	// functionContentType defines file content type (plaintext or base64)
	FunctionContentType string `json:"functionContentType"`

	// size defines as the size of a function pertaining to memory and cpu only. Values can be any one of these S, M, L, XL
	Size string `json:"size"`

	// runtime is the programming language used for a function e.g. nodejs8
	Runtime string `json:"runtime"`

	// timeout defines maximum duration alloted to a function to complete its execution, defaults to 180s
	Timeout int32 `json:"timeout,omitempty"`

	// deps defines the dependencies for a function
	Deps string `json:"deps,omitempty"`

	// envs defines an array of key value pairs need to be used as env variable for a function
	Env []v1.EnvVar `json:"env,omitempty"`
}

// TemplateKind defines the type of BuildTemplate used by the build.
type FunctionCondition string

const (
	// Indicates that function has an unknown condition.
	FunctionConditionUnknown FunctionCondition = "Unknown"
	// Indicates that function has a running condition.
	FunctionConditionRunning FunctionCondition = "Running"
	// Indicates that function has an Building condition. It waits for the Build Pod get the status completed.
	FunctionConditionBuilding FunctionCondition = "Building"
	// Indicates that function has an error condition. Either the Knative Build or the Serving failed.
	FunctionConditionError FunctionCondition = "Error"
	// Indicates that function has a Deploying condition. The knative service is not is status ready yet.
	FunctionConditionDeploying FunctionCondition = "Deploying"
	// Indicates that there is a new image and function is being updated.
	FunctionConditionUpdating FunctionCondition = "Updating"
)

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	Condition FunctionCondition `json:"condition,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Function is the Schema for the functions API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".spec.size",description="Size defines as the size of a function pertaining to memory and cpu only. Values can be any one of these S M L XL)"
// +kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".spec.runtime",description="Runtime is the programming language used for a function e.g. nodejs8"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.condition",description="Check if the function is ready"
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
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
