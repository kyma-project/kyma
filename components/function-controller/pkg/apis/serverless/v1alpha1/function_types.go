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
	// Function defines the content of a function
	Function string `json:"function"`

	// FunctionContentType defines file content type (plaintext or base64)
	FunctionContentType string `json:"functionContentType"`

	// Size defines as the size of a function pertaining to memory and cpu only. Values can be any one of these S, M, L, XL
	Size string `json:"size"`

	// Runtime is the programming language used for a function e.g. nodejs8
	Runtime string `json:"runtime"`

	// Timeout defines maximum duration alloted to a function to complete its execution, defaults to 180s
	Timeout int32 `json:"timeout,omitempty"`

	// Deps defines the dependencies for a function
	Deps string `json:"deps,omitempty"`

	// Env defines an array of key value pairs need to be used as env variable for a function
	Env []v1.EnvVar `json:"env,omitempty"`
}

// ConditionType defines condition of function.
type ConditionType string

const (
	ConditionTypeError        ConditionType = "Error"
	ConditionTypeInitialized  ConditionType = "Initialized"
	ConditionTypeImageCreated ConditionType = "ImageCreated"
	ConditionTypeDeploying    ConditionType = "Deploying"
	ConditionTypeDeployed     ConditionType = "Deployed"
)

type ConditionReason string

const (
	ConditionReasonUnknown                ConditionReason = "Unknown"
	ConditionReasonCreateConfigFailed     ConditionReason = "CreateConfigFailed"
	ConditionReasonCreateConfigSucceeded  ConditionReason = "CreateConfigSucceeded"
	ConditionReasonGetConfigFailed        ConditionReason = "GetConfigFailed"
	ConditionReasonUpdateConfigFailed     ConditionReason = "UpdateConfigFailed"
	ConditionReasonUpdateConfigSucceeded  ConditionReason = "UpdateConfigSucceeded"
	ConditionReasonUpdateRuntimeConfig    ConditionReason = "UpdateRuntimeConfig"
	ConditionReasonBuildFailed            ConditionReason = "BuildFailed"
	ConditionReasonBuildSucceeded         ConditionReason = "BuildSucceeded"
	ConditionReasonDeployFailed           ConditionReason = "DeployFailed"
	ConditionReasonDeploySucceeded        ConditionReason = "DeploySucceeded"
	ConditionReasonCreateServiceSucceeded ConditionReason = "CreateServiceSucceeded"
	ConditionReasonUpdateServiceSucceeded ConditionReason = "UpdateServiceSucceeded"
)

type StatusPhase string

const (
	FunctionPhaseInitializing StatusPhase = "Initializing"
	FunctionPhaseBuilding     StatusPhase = "Building"
	FunctionPhaseDeploying    StatusPhase = "Deploying"
	FunctionPhaseRunning      StatusPhase = "Running"
	FunctionPhaseFailed       StatusPhase = "Failed"
)

type Condition struct {
	Type               ConditionType   `json:"type,omitempty"`
	LastTransitionTime metav1.Time     `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason `json:"reason,omitempty"`
	Message            string          `json:"message,omitempty"`
}

// FunctionStatus defines the observed state of FuncSONPath: .status.phase
type FunctionStatus struct {
	Phase              StatusPhase `json:"phase,omitempty"`
	Conditions         []Condition `json:"conditions,omitempty"`
	ObservedGeneration int64       `json:"observedGeneration,omitempty"`
	ImageTag           string      `json:"imageTag,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Function is the Schema for the functions API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Size",type="string",JSONPath=".spec.size",description="Size defines as the size of a function pertaining to memory and cpu only. Values can be any one of these S M L XL)"
// +kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".spec.runtime",description="Runtime is the programming language used for a function e.g. nodejs8"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase",description="Shows actual function phase, on of: Initializing, Building, Deploying, Running, Error"
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FunctionSpec   `json:"spec,omitempty"`
	Status            FunctionStatus `json:"status,omitempty"`
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
