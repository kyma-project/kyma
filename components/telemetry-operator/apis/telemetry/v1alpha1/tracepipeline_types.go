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
	// Output configures the trace receiver of a TracePipeline.
	Output TracePipelineOutput `json:"output,omitempty"`
}

type TracePipelineOutput struct {
	// Otlp defines an output using the OpenTelmetry protocol.
	Otlp *OtlpOutput `json:"otlp,omitempty"`
}

type OtlpOutput struct {
	// Protocol defines the OTLP protocol (http or grpc).
	Protocol string `json:"protocol,omitempty"`
	// Endpoint defines the host and port (<host>:<port>) of an OTLP endpoint.
	Endpoint ValueType `json:"endpoint,omitempty"`
	// Authentication defines authentication options for the OTLP output
	Authentication *AuthenticationOptions `json:"authentication,omitempty"`
}

type AuthenticationOptions struct {
	// Basic contains credentials for HTTP basic auth
	Basic *BasicAuthOptions `json:"basic,omitempty"`
}

type BasicAuthOptions struct {
	// User contains the basic auth username or a secret reference
	User ValueType `json:"user,omitempty"`
	// Password contains the basic auth password or a secret reference
	Password ValueType `json:"password,omitempty"`
}

func (b *BasicAuthOptions) IsDefined() bool {
	return b.User.IsDefined() && b.Password.IsDefined()
}

// TracePipelineStatus defines the observed state of TracePipeline
type TracePipelineStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// TracePipeline is the Schema for the tracepipelines API
type TracePipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TracePipelineSpec   `json:"spec,omitempty"`
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
