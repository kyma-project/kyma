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

// LogPipelineSpec defines the desired state of LogPipeline
type LogPipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Input     Input               `json:"input,omitempty"`
	Filters   []Filter            `json:"filters,omitempty"`
	Output    Output              `json:"output,omitempty"`
	Files     []FileMount         `json:"files,omitempty"`
	Variables []VariableReference `json:"variables,omitempty"`
}

// Input describes a Fluent Bit input configuration section
type Input struct {
	Application ApplicationInput `json:"application,omitempty"`
}

// ApplicationInput is the default type of Input that handles application logs
type ApplicationInput struct {
	IncludeSystemNamespaces bool     `json:"includeSystemNamespaces,omitempty"`
	Namespaces              []string `json:"namespaces,omitempty"`
	ExcludeNamespaces       []string `json:"excludeNamespaces,omitempty"`
	Containers              []string `json:"containers,omitempty"`
	ExcludeContainers       []string `json:"excludeContainers,omitempty"`
	// KeepAnnotations indicates whether to keep all Kubernetes annotations. The default is false.
	KeepAnnotations bool `json:"keepAnnotations,omitempty"`
	// DropLabels indicates whether to drop all Kubernetes labels. The default is false.
	DropLabels bool `json:"dropLabels,omitempty"`
}

func (a ApplicationInput) HasSelectors() bool {
	return len(a.Namespaces) > 0 ||
		len(a.ExcludeNamespaces) > 0 ||
		len(a.Containers) > 0 ||
		len(a.ExcludeContainers) > 0
}

// Filter describes a Fluent Bit filter configuration
type Filter struct {
	Custom string `json:"custom,omitempty"`
}

// LokiOutput describes a Fluent Bit Loki output configuration
type LokiOutput struct {
	URL        ValueType         `json:"url,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	RemoveKeys []string          `json:"removeKeys,omitempty"`
}

// HttpOutput describes a Fluent Bit HTTP output configuration
type HTTPOutput struct {
	Host      ValueType `json:"host,omitempty"`
	User      ValueType `json:"user,omitempty"`
	Password  ValueType `json:"password,omitempty"`
	URI       string    `json:"uri,omitempty"`
	Port      string    `json:"port,omitempty"`
	Compress  string    `json:"compress,omitempty"`
	Format    string    `json:"format,omitempty"`
	TLSConfig TLSConfig `json:"tls,omitempty"`
	Dedot     bool      `json:"dedot,omitempty"`
}

type TLSConfig struct {
	Disabled                  bool `json:"disabled,omitempty"`
	SkipCertificateValidation bool `json:"skipCertificateValidation,omitempty"`
}

// Output describes a Fluent Bit output configuration section
type Output struct {
	Custom string     `json:"custom,omitempty"`
	HTTP   HTTPOutput `json:"http,omitempty"`
	Loki   LokiOutput `json:"grafana-loki,omitempty"`
}

// FileMount provides file content to be consumed by a LogPipeline configuration
type FileMount struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

// VariableReference references a Kubernetes secret that should be provided as environment variable to Fluent Bit
type VariableReference struct {
	Name      string        `json:"name,omitempty"`
	ValueFrom ValueFromType `json:"valueFrom,omitempty"`
}

type ValueType struct {
	Value     string        `json:"value,omitempty"`
	ValueFrom ValueFromType `json:"valueFrom,omitempty"`
}

func (v *ValueType) IsDefined() bool {
	return v.Value != "" || v.ValueFrom.IsSecretRef()
}

func (v *ValueFromType) IsSecretRef() bool {
	return v.SecretKey.Name != "" && v.SecretKey.Key != ""
}

type ValueFromType struct {
	SecretKey SecretKeyRef `json:"secretKeyRef,omitempty"`
}

type SecretKeyRef struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key,omitempty"`
}

type LogPipelineConditionType string

// These are the valid statuses of LogPipeline.
const (
	LogPipelinePending LogPipelineConditionType = "Pending"
	LogPipelineRunning LogPipelineConditionType = "Running"
)

const (
	FluentBitDSRestartedReason        = "FluentBitDaemonSetRestarted"
	FluentBitDSRestartCompletedReason = "FluentBitDaemonSetRestartCompleted"
	SecretsNotPresent                 = "OneORMoreSecretsAreNotPresent"
)

// LogPipelineCondition contains details for the current condition of this LogPipeline
type LogPipelineCondition struct {
	LastTransitionTime metav1.Time              `json:"lastTransitionTime,omitempty"`
	Reason             string                   `json:"reason,omitempty"`
	Type               LogPipelineConditionType `json:"type,omitempty"`
}

// LogPipelineStatus defines the observed state of LogPipeline
type LogPipelineStatus struct {
	Conditions      []LogPipelineCondition `json:"conditions,omitempty"`
	UnsupportedMode bool                   `json:"unsupportedMode,omitempty"`
}

func NewLogPipelineCondition(reason string, condType LogPipelineConditionType) *LogPipelineCondition {
	return &LogPipelineCondition{
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Type:               condType,
	}
}

func (lps *LogPipelineStatus) GetCondition(condType LogPipelineConditionType) *LogPipelineCondition {
	for cond := range lps.Conditions {
		if lps.Conditions[cond].Type == condType {
			return &lps.Conditions[cond]
		}
	}
	return nil
}

func (lps *LogPipelineStatus) SetCondition(cond LogPipelineCondition) {
	currentCond := lps.GetCondition(cond.Type)
	if currentCond != nil && currentCond.Reason == cond.Reason {
		return
	}
	if currentCond != nil {
		cond.LastTransitionTime = currentCond.LastTransitionTime
	}
	newConditions := filterOutCondition(lps.Conditions, cond.Type)
	lps.Conditions = append(newConditions, cond)
}

func filterOutCondition(conditions []LogPipelineCondition, condType LogPipelineConditionType) []LogPipelineCondition {
	var newConditions []LogPipelineCondition
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
// +kubebuilder:printcolumn:name="Unsupported-Mode",type=boolean,JSONPath=`.status.unsupportedMode`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// LogPipeline is the Schema for the logpipelines API
type LogPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogPipelineSpec   `json:"spec,omitempty"`
	Status LogPipelineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
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
