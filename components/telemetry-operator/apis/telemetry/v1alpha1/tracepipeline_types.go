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

// TracePipelineSpec defines the desired state of TracePipeline
type TracePipelineSpec struct {
	// Configures the trace receiver of a TracePipeline.
	Output TracePipelineOutput `json:"output"`
}

// TracePipelineOutput defines the output configuration section.
type TracePipelineOutput struct {
	// Defines an output using the OpenTelmetry protocol.
	Otlp *OtlpOutput `json:"otlp"`
}

type Header struct {
	// Defines the header name
	Name string `json:"name"`
	// Defines the header value
	ValueType `json:",inline"`
}

type OtlpOutput struct {
	// Defines the OTLP protocol (http or grpc).
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default:=grpc
	// +kubebuilder:validation:Enum=grpc;http
	Protocol string `json:"protocol,omitempty"`
	// Defines the host and port (<host>:<port>) of an OTLP endpoint.
	// +kubebuilder:validation:Required
	Endpoint ValueType `json:"endpoint"`
	// Defines authentication options for the OTLP output
	Authentication *AuthenticationOptions `json:"authentication,omitempty"`
	// Custom headers to be added to outgoing HTTP or GRPC requests
	Headers []Header `json:"headers,omitempty"`
}

type AuthenticationOptions struct {
	// Contains credentials for HTTP basic auth
	Basic *BasicAuthOptions `json:"basic,omitempty"`
}

type BasicAuthOptions struct {
	// Contains the basic auth username or a secret reference
	// +kubebuilder:validation:Required
	User ValueType `json:"user"`
	// Contains the basic auth password or a secret reference
	// +kubebuilder:validation:Required
	Password ValueType `json:"password"`
}

func (b *BasicAuthOptions) IsDefined() bool {
	return b.User.IsDefined() && b.Password.IsDefined()
}

type TracePipelineConditionType string

// These are the valid statuses of TracePipeline.
const (
	TracePipelinePending TracePipelineConditionType = "Pending"
	TracePipelineRunning TracePipelineConditionType = "Running"
)

// Contains details for the current condition of this TracePipeline
type TracePipelineCondition struct {
	LastTransitionTime metav1.Time                `json:"lastTransitionTime,omitempty"`
	Reason             string                     `json:"reason,omitempty"`
	Type               TracePipelineConditionType `json:"type,omitempty"`
}

// Defines the observed state of TracePipeline
type TracePipelineStatus struct {
	Conditions []TracePipelineCondition `json:"conditions,omitempty"`
}

func NewTracePipelineCondition(reason string, condType TracePipelineConditionType) *TracePipelineCondition {
	return &TracePipelineCondition{
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Type:               condType,
	}
}

func (tps *TracePipelineStatus) GetCondition(condType TracePipelineConditionType) *TracePipelineCondition {
	for cond := range tps.Conditions {
		if tps.Conditions[cond].Type == condType {
			return &tps.Conditions[cond]
		}
	}
	return nil
}

func (tps *TracePipelineStatus) HasCondition(condition TracePipelineConditionType) bool {
	return tps.GetCondition(condition) != nil
}

func (tps *TracePipelineStatus) SetCondition(cond TracePipelineCondition) {
	currentCond := tps.GetCondition(cond.Type)
	if currentCond != nil && currentCond.Reason == cond.Reason {
		return
	}
	if currentCond != nil {
		cond.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterCondition(tps.Conditions, cond.Type)
	tps.Conditions = append(newConditions, cond)
}

func filterCondition(conditions []TracePipelineCondition, condType TracePipelineConditionType) []TracePipelineCondition {
	var newConditions []TracePipelineCondition
	for _, cond := range conditions {
		if cond.Type == condType {
			continue
		}
		newConditions = append(newConditions, cond)
	}
	return newConditions
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[-1].type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// TracePipeline is the Schema for the tracepipelines API
type TracePipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of TracePipeline
	Spec TracePipelineSpec `json:"spec,omitempty"`
	// Shows the observed state of the TracePipeline
	Status TracePipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TracePipelineList contains a list of TracePipeline
type TracePipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TracePipeline `json:"items"`
}

//nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&TracePipeline{}, &TracePipelineList{})
}
