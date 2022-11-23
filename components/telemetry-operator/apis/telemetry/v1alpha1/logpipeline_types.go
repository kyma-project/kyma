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

// LogPipelineSpec defines the desired state of LogPipeline
type LogPipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Describes a log input for a LogPipeline.
	Input   Input    `json:"input,omitempty"`
	Filters []Filter `json:"filters,omitempty"`
	// Describes a Fluent Bit output configuration section.
	Output    Output        `json:"output,omitempty"`
	Files     []FileMount   `json:"files,omitempty"`
	Variables []VariableRef `json:"variables,omitempty"`
}

// Input describes a log input for a LogPipeline.
type Input struct {
	// Configures in more detail from which containers application logs are enabled as input.
	Application ApplicationInput `json:"application,omitempty"`
}

// ApplicationInput specifies the default type of Input that handles application logs from runtime containers. It configures in more detail from which containers logs are selected as input.
type ApplicationInput struct {
	// Describes whether application logs from specific Namespaces are selected. The options are mutually exclusive. System Namespaces are excluded by default from the collection.
	Namespaces InputNamespaces `json:"namespaces,omitempty"`
	// Describes whether application logs from specific containers are selected. The options are mutually exclusive.
	Containers InputContainers `json:"containers,omitempty"`
	// Defines whether to keep all Kubernetes annotations. The default is false.
	KeepAnnotations bool `json:"keepAnnotations,omitempty"`
	// Defines whether to drop all Kubernetes labels. The default is false.
	DropLabels bool `json:"dropLabels,omitempty"`
}

// InputNamespaces describes whether application logs from specific Namespaces are selected. The options are mutually exclusive. System Namespaces are excluded by default from the collection.
type InputNamespaces struct {
	// Include only the container logs of the specified Namespace names.
	Include []string `json:"include,omitempty"`
	// Exclude the container logs of the specified Namespace names.
	Exclude []string `json:"exclude,omitempty"`
	// Describes to include the container logs of the system Namespaces like kube-system, istio-system, and kyma-system.
	System bool `json:"system,omitempty"`
}

// InputContainers describes whether application logs from specific containers are selected. The options are mutually exclusive.
type InputContainers struct {
	// Specifies to include only the container logs with the specified container names.
	Include []string `json:"include,omitempty"`
	// Specifies to exclude only the container logs with the specified container names.
	Exclude []string `json:"exclude,omitempty"`
}

// Describes a filtering option on the logs of the pipeline.
type Filter struct {
	// Custom filter definition in the Fluent Bit syntax. Note: If you use a `custom` filter, you put the LogPipeline in unsupported mode.
	Custom string `json:"custom,omitempty"`
}

// LokiOutput configures an output to the Kyma-internal Loki instance.
type LokiOutput struct {
	URL        ValueType         `json:"url,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	RemoveKeys []string          `json:"removeKeys,omitempty"`
}

// HTTPOutput configures an HTTP-based output compatible with the Fluent Bit HTTP output plugin.
type HTTPOutput struct {
	// Defines the host of the HTTP receiver.
	Host ValueType `json:"host,omitempty"`
	// Defines the basic auth user.
	User ValueType `json:"user,omitempty"`
	// Defines the basic auth password.
	Password ValueType `json:"password,omitempty"`
	// Defines the URI of the HTTP receiver. Default is "/".
	URI string `json:"uri,omitempty"`
	// Defines the port of the HTTP receiver. Default is 443.
	Port string `json:"port,omitempty"`
	// Defines the compression algorithm to use.
	Compress string `json:"compress,omitempty"`
	// Defines the log encoding to be used. Default is json.
	Format string `json:"format,omitempty"`
	// Defines TLS settings for the HTTP connection.
	TLSConfig TLSConfig `json:"tls,omitempty"`
	// Enables de-dotting of Kubernetes labels and annotations for compatibility with ElasticSearch based backends. Dots (.) will be replaced by underscores (_).
	Dedot bool `json:"dedot,omitempty"`
}

type TLSConfig struct {
	// Disable TLS.
	Disabled bool `json:"disabled,omitempty"`
	// Disable TLS certificate validation.
	SkipCertificateValidation bool `json:"skipCertificateValidation,omitempty"`
}

// Output describes a Fluent Bit output configuration section.
type Output struct {
	// Defines a custom output in the Fluent Bit syntax. Note: If you use a `custom` output, you put the LogPipeline in unsupported mode.
	Custom string `json:"custom,omitempty"`
	// Configures an HTTP-based output compatible with the Fluent Bit HTTP output plugin.
	HTTP *HTTPOutput `json:"http,omitempty"`
	// Configures an output to the Kyma-internal Loki instance. Note: This output is considered legacy and is only provided for backwards compatibility with the in-cluster Loki instance. It might not be compatible with latest Loki versions. For integration with a Loki-based system, use the `custom` output with name `loki` instead.
	Loki *LokiOutput `json:"grafana-loki,omitempty"`
}

func (o *Output) IsCustomDefined() bool {
	return o.Custom != ""
}

func (o *Output) IsHTTPDefined() bool {
	return o.HTTP != nil && o.HTTP.Host.IsDefined()
}

func (o *Output) IsLokiDefined() bool {
	return o.Loki != nil && o.Loki.URL.IsDefined()
}

func (o *Output) IsAnyDefined() bool {
	return o.pluginCount() > 0
}

func (o *Output) IsSingleDefined() bool {
	return o.pluginCount() == 1
}

func (o *Output) pluginCount() int {
	plugins := 0
	if o.IsCustomDefined() {
		plugins++
	}
	if o.IsHTTPDefined() {
		plugins++
	}
	if o.IsLokiDefined() {
		plugins++
	}
	return plugins
}

// Provides file content to be consumed by a LogPipeline configuration
type FileMount struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

// References a Kubernetes secret that should be provided as environment variable to Fluent Bit
type VariableRef struct {
	Name      string          `json:"name,omitempty"`
	ValueFrom ValueFromSource `json:"valueFrom,omitempty"`
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

// LogPipelineStatus shows the observed state of the LogPipeline
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

	// Defines the desired state of LogPipeline
	Spec LogPipelineSpec `json:"spec,omitempty"`
	// Shows the observed state of the LogPipeline
	Status LogPipelineStatus `json:"status,omitempty"`
}

// ContainsCustomPlugin returns true if the pipeline contains any custom filters or outputs
func (l *LogPipeline) ContainsCustomPlugin() bool {
	for _, filter := range l.Spec.Filters {
		if filter.Custom != "" {
			return true
		}
	}
	return l.Spec.Output.IsCustomDefined()
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
