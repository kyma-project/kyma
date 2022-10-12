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

package v1alpha2

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Runtime specifies the name of the Function's runtime.
type Runtime string

const (
	Python39 Runtime = "python39"
	NodeJs12 Runtime = "nodejs12"
	NodeJs14 Runtime = "nodejs14"
	NodeJs16 Runtime = "nodejs16"
)

type FunctionType string

const (
	FunctionTypeInline FunctionType = "inline"
	FunctionTypeGit    FunctionType = "git"
)

type Source struct {
	// +optional
	GitRepository *GitRepositorySource `json:"gitRepository,omitempty"`
	// +optional
	Inline *InlineSource `json:"inline,omitempty"`
}

type InlineSource struct {
	Source string `json:"source"`
	//+optional
	Dependencies string `json:"dependencies,omitempty"`
}

type GitRepositorySource struct {
	// +kubebuilder:validation:Required

	// URL is the address of GIT repository
	URL string `json:"url"`

	// Auth is the optional definition of authentication that should be used for repository operations
	// +optional
	Auth *RepositoryAuth `json:"auth,omitempty"`

	Repository `json:",inline"`
}

// RepositoryAuth defines authentication method used for repository operations
type RepositoryAuth struct {
	// Type is the type of authentication
	Type RepositoryAuthType `json:"type"`

	// +kubebuilder:validation:Required

	// SecretName is the name of Kubernetes Secret containing credentials used for authentication
	SecretName string `json:"secretName"`
}

// RepositoryAuthType is the enum of available authentication types
// +kubebuilder:validation:Enum=basic;key
type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey RepositoryAuthType = "key"
)

type ConfigMapRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Template struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type ResourceRequirements struct {
	// +optional
	// Set resources with predefined values.
	Profile string `json:"profile,omitempty"`

	// +optional
	//DEPRECATED: use templates.functionPod.spec.resources or
	//templates.buildJob.spec.resources to configure resources
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ScaleConfig struct {
	// +kubebuilder:validation:Minimum:=1
	MinReplicas *int32 `json:"minReplicas"`

	// +kubebuilder:validation:Minimum:=1
	MaxReplicas *int32 `json:"maxReplicas"`
}

type ResourceConfiguration struct {
	// +optional
	Build *ResourceRequirements `json:"build,omitempty"`
	// +optional
	Function *ResourceRequirements `json:"function,omitempty"`
}

type PodSpecTemplate struct {
	// Compute Resources required by container created by function controller.
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`

	// List of environment variables to set in the container.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []v1.EnvVar `json:"env,omitempty"`

	// Pod volumes to mount into the container's filesystem.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
}

type MetadataTemplate struct {
	// Additional labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Additional annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type PodTemplate struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	Metadata *MetadataTemplate `json:"metadata,omitempty"`

	// Additional specification for pod created by function-controller
	//+optional
	Spec *PodSpecTemplate `json:"spec,omitempty"`

	// Configuration of available volumes for pod
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
}

type Templates struct {
	// Additional specification for build job's pod
	// +optional
	BuildJob *PodTemplate `json:"buildJob,omitempty"`

	// Additional specification for function's pod
	// +optional
	FunctionPod *PodTemplate `json:"functionPod,omitempty"`
}

const (
	FunctionResourcesPresetLabel = "serverless.kyma-project.io/function-resources-preset"
	BuildResourcesPresetLabel    = "serverless.kyma-project.io/build-resources-preset"
)

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	Runtime Runtime `json:"runtime"`

	//+optional
	CustomRuntimeConfiguration *ConfigMapRef `json:"customRuntimeConfiguration,omitempty"`

	// +optional
	RuntimeImageOverride string `json:"runtimeImageOverride,omitempty"`

	Source Source `json:"source"`

	//DEPRECATED: use spec.templates.functionPod.env or spec.templates.buildJob.env
	//Env defines an array of key value pairs need to be used as env variable for a function
	// +optional
	Env []v1.EnvVar `json:"env,omitempty"`

	// +optional
	ResourceConfiguration *ResourceConfiguration `json:"resourceConfiguration,omitempty"`

	// +optional
	ScaleConfig *ScaleConfig `json:"scaleConfig,omitempty"`

	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	//DEPRECATED: use templates.functionPod.metadata for function's pod
	//templates.buildJob.metadata for function's build job
	Template *Template `json:"template,omitempty"`

	// +optional
	// Additional configuration of function's created pods
	Templates *Templates `json:"templates,omitempty"`
}

// TODO: Status related things needs to be developed.
type ConditionType string

const (
	ConditionRunning            ConditionType = "Running"
	ConditionConfigurationReady ConditionType = "ConfigurationReady"
	ConditionBuildReady         ConditionType = "BuildReady"
)

type ConditionReason string

const (
	ConditionReasonConfigMapCreated               ConditionReason = "ConfigMapCreated"
	ConditionReasonConfigMapUpdated               ConditionReason = "ConfigMapUpdated"
	ConditionReasonSourceUpdated                  ConditionReason = "SourceUpdated"
	ConditionReasonSourceUpdateFailed             ConditionReason = "SourceUpdateFailed"
	ConditionReasonJobFailed                      ConditionReason = "JobFailed"
	ConditionReasonJobCreated                     ConditionReason = "JobCreated"
	ConditionReasonJobUpdated                     ConditionReason = "JobUpdated"
	ConditionReasonJobRunning                     ConditionReason = "JobRunning"
	ConditionReasonJobsDeleted                    ConditionReason = "JobsDeleted"
	ConditionReasonJobFinished                    ConditionReason = "JobFinished"
	ConditionReasonDeploymentCreated              ConditionReason = "DeploymentCreated"
	ConditionReasonDeploymentUpdated              ConditionReason = "DeploymentUpdated"
	ConditionReasonDeploymentFailed               ConditionReason = "DeploymentFailed"
	ConditionReasonDeploymentWaiting              ConditionReason = "DeploymentWaiting"
	ConditionReasonDeploymentReady                ConditionReason = "DeploymentReady"
	ConditionReasonServiceCreated                 ConditionReason = "ServiceCreated"
	ConditionReasonServiceUpdated                 ConditionReason = "ServiceUpdated"
	ConditionReasonHorizontalPodAutoscalerCreated ConditionReason = "HorizontalPodAutoscalerCreated"
	ConditionReasonHorizontalPodAutoscalerUpdated ConditionReason = "HorizontalPodAutoscalerUpdated"
	ConditionReasonMinReplicasNotAvailable        ConditionReason = "MinReplicasNotAvailable"
)

type Condition struct {
	Type               ConditionType      `json:"type,omitempty"`
	Status             v1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason    `json:"reason,omitempty"`
	Message            string             `json:"message,omitempty"`
}

type Repository struct {
	BaseDir   string `json:"baseDir,omitempty"`
	Reference string `json:"reference,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	Runtime              Runtime     `json:"runtime,omitempty"`
	Conditions           []Condition `json:"conditions,omitempty"`
	Repository           `json:",inline,omitempty"`
	Replicas             int32  `json:"replicas,omitempty"`
	PodSelector          string `json:"podSelector,omitempty"`
	Commit               string `json:"commit,omitempty"`
	RuntimeImageOverride string `json:"runtimeImageOverride,omitempty"`
}

const (
	FunctionNameLabel                    = "serverless.kyma-project.io/function-name"
	FunctionManagedByLabel               = "serverless.kyma-project.io/managed-by"
	FunctionControllerValue              = "function-controller"
	FunctionUUIDLabel                    = "serverless.kyma-project.io/uuid"
	FunctionResourceLabel                = "serverless.kyma-project.io/resource"
	FunctionResourceLabelDeploymentValue = "deployment"
	FunctionResourceLabelUserValue       = "user"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.podSelector
//+kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='ConfigurationReady')].status"
//+kubebuilder:printcolumn:name="Built",type="string",JSONPath=".status.conditions[?(@.type=='BuildReady')].status"
//+kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
//+kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".spec.runtime"
//+kubebuilder:printcolumn:name="Version",type="integer",JSONPath=".metadata.generation"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FunctionSpec   `json:"spec,omitempty"`
	Status FunctionStatus `json:"status,omitempty"`
}

func (f *Function) TypeOf(t FunctionType) bool {
	switch t {

	case FunctionTypeInline:
		return f.Spec.Source.Inline != nil

	case FunctionTypeGit:
		return f.Spec.Source.GitRepository != nil

	default:
		return false
	}
}

//+kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

// nolint
func init() {
	SchemeBuilder.Register(
		&Function{},
		&FunctionList{},
	)
}

func (s *FunctionStatus) Condition(c ConditionType) *Condition {
	for _, cond := range s.Conditions {
		if cond.Type == c {
			return &cond
		}
	}
	return nil
}

func (c *Condition) IsTrue() bool {
	return c.Status == v1.ConditionTrue
}

func (l *Condition) Equal(r *Condition) bool {
	if l == nil && r == nil {
		return true
	}

	if l.Type != r.Type ||
		l.Status != r.Status ||
		l.Reason != r.Reason ||
		l.Message != r.Message ||
		!l.LastTransitionTime.Equal(&r.LastTransitionTime) {
		return false
	}
	return true
}

func (p *PodTemplate) HasVolumes() bool {
	if p == nil || p.Volumes == nil {
		return false
	}
	return true
}

func (p *PodTemplate) HasVolumeMounts() bool {
	if p == nil || p.Spec == nil || p.Spec.VolumeMounts == nil {
		return false
	}
	return true
}
