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
	NodeJs14 Runtime = "nodejs14"
	NodeJs16 Runtime = "nodejs16"
)

type FunctionType string

const (
	FunctionTypeInline FunctionType = "inline"
	FunctionTypeGit    FunctionType = "git"
)

type Source struct {
	// GitRepository defines Function as git-sourced. Can't be used at the same time with Inline.
	// +optional
	GitRepository *GitRepositorySource `json:"gitRepository,omitempty"`

	// Inline defines Function as the inline Function. Can't be used at the same time with GitRepository.
	// +optional
	Inline *InlineSource `json:"inline,omitempty"`
}

type InlineSource struct {
	// Source provides the Function's full source code.
	Source string `json:"source"`

	// Dependencies specifies the Function's dependencies.
	//+optional
	Dependencies string `json:"dependencies,omitempty"`
}

type GitRepositorySource struct {
	// +kubebuilder:validation:Required

	// URL provides the address to the Git repository with the Function's code and dependencies.
	// Depending on whether the repository is public or private and what authentication method is used to access it,
	// the URL must start with the `http(s)`, `git`, or `ssh` prefix.
	URL string `json:"url"`

	// Auth specifies that you must authenticate to the Git repository. Required for SSH.
	// +optional
	Auth *RepositoryAuth `json:"auth,omitempty"`

	Repository `json:",inline"`
}

// RepositoryAuth defines authentication method used for repository operations
type RepositoryAuth struct {
	// RepositoryAuthType defines if you must authenticate to the repository with a password or token (`basic`),
	// or an SSH key (`key`). For SSH, this parameter must be set to `key`.
	Type RepositoryAuthType `json:"type"`

	// +kubebuilder:validation:Required

	// SecretName specifies the name of the Secret with credentials used by the Function Controller
	// to authenticate to the Git repository in order to fetch the Function's source code and dependencies.
	// This Secret must be stored in the same Namespace as the Function CR.
	SecretName string `json:"secretName"`
}

// RepositoryAuthType is the enum of available authentication types
// +kubebuilder:validation:Enum=basic;key
type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey RepositoryAuthType = "key"
)

type Template struct {
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type ResourceRequirements struct {
	// Profile defines name of predefined set of values of resource. Can't be used at the same time with Resources.
	// +optional
	Profile string `json:"profile,omitempty"`

	// Resources defines amount of resources available for the Pod to use. Can't be used at the same time with Profile.
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ScaleConfig struct {
	// MinReplicas defines the minimum number of Function's Pods to run at a time.
	// +kubebuilder:validation:Minimum:=1
	MinReplicas *int32 `json:"minReplicas"`

	// MaxReplicas defines the maximum number of Function's Pods to run at a time.
	// +kubebuilder:validation:Minimum:=1
	MaxReplicas *int32 `json:"maxReplicas"`
}

type ResourceConfiguration struct {
	// Build specifies resources requested by the build Job's Pod.
	// +optional
	Build *ResourceRequirements `json:"build,omitempty"`

	// Function specifies resources requested by the Function's Pod.
	// +optional
	Function *ResourceRequirements `json:"function,omitempty"`
}

type SecretMount struct {
	// SecretName specifies name of the Secret in the Function's Namespace to use.
	// +kubebuilder:validation:Required
	SecretName string `json:"secretName"`

	// MountPath specifies path within the container at which the Secret should be mounted.
	// +kubebuilder:validation:Required
	MountPath string `json:"mountPath"`
}

const (
	FunctionResourcesPresetLabel = "serverless.kyma-project.io/function-resources-preset"
	BuildResourcesPresetLabel    = "serverless.kyma-project.io/build-resources-preset"
)

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// Runtime specifies the runtime of the Function. The available values are `nodejs14`, `nodejs16`, and `python39`.
	Runtime Runtime `json:"runtime"`

	// RuntimeImageOverride specifies the runtimes image which must be used instead of the default one.
	// +optional
	RuntimeImageOverride string `json:"runtimeImageOverride,omitempty"`

	// Source contains the Function's specification.
	Source Source `json:"source"`

	// Env specifies an array of key-value pairs to be used as environment variables for the Function.
	// You can define values as static strings or reference values from ConfigMaps or Secrets.
	Env []v1.EnvVar `json:"env,omitempty"`

	// ResourceConfiguration specifies resources requested by Function and build Job.
	// +optional
	ResourceConfiguration *ResourceConfiguration `json:"resourceConfiguration,omitempty"`

	// ScaleConfig defines minimum and maximum number of Function's Pods to run at a time.
	// When it is configured, a HorizontalPodAutoscaler will be deployed and will control the Replicas field
	// to scale Function based on the CPU utilisation.
	// +optional
	ScaleConfig *ScaleConfig `json:"scaleConfig,omitempty"`

	// Replicas defines the exact number of Function's Pods to run at a time.
	// If ScaleConfig is configured, or if Function is targeted by an external scaler,
	// then the Replicas field is used by the relevant HorizontalPodAutoscaler to control the number of active replicas.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Template *Template `json:"template,omitempty"`

	// SecretMounts specifies Secrets to mount into the Function's container filesystem.
	SecretMounts []SecretMount `json:"secretMounts,omitempty"`
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
	// BaseDir specifies the relative path to the Git directory that contains the source code
	// from which the Function is built.
	BaseDir string `json:"baseDir,omitempty"`

	// Reference specifies either the branch name, tag or the commit revision from which the Function Controller
	// automatically fetches the changes in the Function's code and dependencies.
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

// A simple code snippet that you can run without provisioning or managing servers.
// It implements the exact business logic you define.
// A Function is based on the Function custom resource (CR) and can be written in either Node.js or Python.
// A Function can perform a business logic of its own. You can also bind it to an instance of a service
// and configure it to be triggered whenever it receives a particular event type from the service
// or a call is made to the service's API.
// Functions are executed only if they are triggered by an event or an API call.
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
