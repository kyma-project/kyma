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
	// Deprecated: Nodejs16 will be removed soon
	NodeJs16 Runtime = "nodejs16"
	NodeJs18 Runtime = "nodejs18"
)

type FunctionType string

const (
	FunctionTypeInline FunctionType = "inline"
	FunctionTypeGit    FunctionType = "git"
)

type Source struct {
	// Defines the Function as git-sourced. Can't be used together with **Inline**.
	// +optional
	GitRepository *GitRepositorySource `json:"gitRepository,omitempty"`

	// Defines the Function as the inline Function. Can't be used together with **GitRepository**.
	// +optional
	Inline *InlineSource `json:"inline,omitempty"`
}

type InlineSource struct {
	// Specifies the Function's full source code.
	Source string `json:"source"`

	// Specifies the Function's dependencies.
	//+optional
	Dependencies string `json:"dependencies,omitempty"`
}

type GitRepositorySource struct {
	// +kubebuilder:validation:Required

	// Specifies the URL of the Git repository with the Function's code and dependencies.
	// Depending on whether the repository is public or private and what authentication method is used to access it,
	// the URL must start with the `http(s)`, `git`, or `ssh` prefix.
	URL string `json:"url"`

	// Specifies the authentication method. Required for SSH.
	// +optional
	Auth *RepositoryAuth `json:"auth,omitempty"`

	Repository `json:",inline"`
}

// RepositoryAuth defines authentication method used for repository operations
type RepositoryAuth struct {
	// Defines the repository authentication method. The value is either `basic` if you use a password or token,
	// or `key` if you use an SSH key.
	Type RepositoryAuthType `json:"type"`

	// +kubebuilder:validation:Required

	// Specifies the name of the Secret with credentials used by the Function Controller
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
	// Deprecated: Use **FunctionSpec.Labels**  to label Function's Pods.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
	// Deprecated: Use **FunctionSpec.Annotations** to annotate Function's Pods.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

type ResourceRequirements struct {
	// Defines the name of the predefined set of values of the resource.
	// Can't be used together with **Resources**.
	// +optional
	Profile string `json:"profile,omitempty"`

	// Defines the amount of resources available for the Pod.
	// Can't be used together with **Profile**.
	// For configuration details, see the [official Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).
	// +optional
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ScaleConfig struct {
	// Defines the minimum number of Function's Pods to run at a time.
	// +kubebuilder:validation:Minimum:=1
	MinReplicas *int32 `json:"minReplicas"`

	// Defines the maximum number of Function's Pods to run at a time.
	// +kubebuilder:validation:Minimum:=1
	MaxReplicas *int32 `json:"maxReplicas"`
}

type ResourceConfiguration struct {
	// Specifies resources requested by the build Job's Pod.
	// +optional
	Build *ResourceRequirements `json:"build,omitempty"`

	// Specifies resources requested by the Function's Pod.
	// +optional
	Function *ResourceRequirements `json:"function,omitempty"`
}

type SecretMount struct {
	// Specifies the name of the Secret in the Function's Namespace.
	// +kubebuilder:validation:Required
	SecretName string `json:"secretName"`

	// Specifies the path within the container where the Secret should be mounted.
	// +kubebuilder:validation:Required
	MountPath string `json:"mountPath"`
}

const (
	FunctionResourcesPresetLabel = "serverless.kyma-project.io/function-resources-preset"
	BuildResourcesPresetLabel    = "serverless.kyma-project.io/build-resources-preset"
)

// Defines the desired state of the Function
type FunctionSpec struct {
	// Specifies the runtime of the Function. The available values are `nodejs16` - deprecated, `nodejs18`, and `python39`.
	Runtime Runtime `json:"runtime"`

	// Specifies the runtime image used instead of the default one.
	// +optional
	RuntimeImageOverride string `json:"runtimeImageOverride,omitempty"`

	// Contains the Function's source code configuration.
	Source Source `json:"source"`

	// Specifies an array of key-value pairs to be used as environment variables for the Function.
	// You can define values as static strings or reference values from ConfigMaps or Secrets.
	// For configuration details, see the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/).
	Env []v1.EnvVar `json:"env,omitempty"`

	// Specifies resources requested by the Function and the build Job.
	// +optional
	ResourceConfiguration *ResourceConfiguration `json:"resourceConfiguration,omitempty"`

	// Defines the minimum and maximum number of Function's Pods to run at a time.
	// When it is configured, a HorizontalPodAutoscaler will be deployed and will control the **Replicas** field
	// to scale the Function based on the CPU utilisation.
	// +optional
	ScaleConfig *ScaleConfig `json:"scaleConfig,omitempty"`

	// Defines the exact number of Function's Pods to run at a time.
	// If **ScaleConfig** is configured, or if the Function is targeted by an external scaler,
	// then the **Replicas** field is used by the relevant HorizontalPodAutoscaler to control the number of active replicas.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Deprecated: Use **Labels** and **Annotations** to label and/or annotate Function's Pods.
	// +optional
	Template *Template `json:"template,omitempty"`

	// Specifies Secrets to mount into the Function's container filesystem.
	SecretMounts []SecretMount `json:"secretMounts,omitempty"`

	// Defines labels used in Deployment's PodTemplate and applied on the Function's runtime Pod.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Defines annotations used in Deployment's PodTemplate and applied on the Function's runtime Pod.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
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
	ConditionReasonServiceFailed                  ConditionReason = "ServiceFailed"
	ConditionReasonHorizontalPodAutoscalerCreated ConditionReason = "HorizontalPodAutoscalerCreated"
	ConditionReasonHorizontalPodAutoscalerUpdated ConditionReason = "HorizontalPodAutoscalerUpdated"
	ConditionReasonMinReplicasNotAvailable        ConditionReason = "MinReplicasNotAvailable"
)

type Condition struct {
	// Specifies the type of the Function's condition.
	Type ConditionType `json:"type,omitempty"`
	// Specifies the status of the condition. The value is either `True`, `False`, or `Unknown`.
	Status v1.ConditionStatus `json:"status"`
	// Specifies the last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Specifies the reason for the condition's last transition.
	Reason ConditionReason `json:"reason,omitempty"`
	// Provides a human-readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type Repository struct {
	// Specifies the relative path to the Git directory that contains the source code
	// from which the Function is built.
	BaseDir string `json:"baseDir,omitempty"`

	// Specifies either the branch name, tag or commit revision from which the Function Controller
	// automatically fetches the changes in the Function's code and dependencies.
	Reference string `json:"reference,omitempty"`
}

// FunctionStatus defines the observed state of the Function
type FunctionStatus struct {
	// Specifies the **Runtime** type of the Function.
	Runtime Runtime `json:"runtime,omitempty"`
	// Specifies an array of conditions describing the status of the parser.
	Conditions []Condition `json:"conditions,omitempty"`
	// Specify the repository which was used to build the function.
	Repository `json:",inline,omitempty"`
	// Specifies the total number of non-terminated Pods targeted by this Function.
	Replicas int32 `json:"replicas,omitempty"`
	// Specifies the Pod selector used to match Pods in the Function's Deployment.
	PodSelector string `json:"podSelector,omitempty"`
	// Specifies the commit hash used to build the Function.
	Commit string `json:"commit,omitempty"`
	// Specifies the image version used to build and run the Function's Pods.
	RuntimeImage string `json:"runtimeImage,omitempty"`
	// Deprecated: Specifies the runtime image version which overrides the **RuntimeImage** status parameter.
	// **RuntimeImageOverride** exists for historical compatibility
	// and should be removed with v1alpha3 version.
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

func (f *Function) EffectiveRuntime() string {
	if f.Spec.RuntimeImageOverride != "" {
		return f.Spec.RuntimeImageOverride
	}
	return string(f.Spec.Runtime)
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

func (f Function) IsUpdating() bool {
	conditions := []ConditionType{ConditionBuildReady, ConditionConfigurationReady, ConditionRunning}
	status := f.Status
	for _, c := range conditions {
		if !status.Condition(c).IsTrue() {
			return true
		}
	}
	return false
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

func (rc *ResourceRequirements) EffectiveResource(defaultProfile string, profiles map[string]v1.ResourceRequirements) v1.ResourceRequirements {
	if rc == nil {
		return profiles[defaultProfile]
	}
	profileResources, found := profiles[rc.Profile]
	if found {
		return profileResources
	}
	if rc.Resources != nil {
		return *rc.Resources
	}
	return profiles[defaultProfile]
}
