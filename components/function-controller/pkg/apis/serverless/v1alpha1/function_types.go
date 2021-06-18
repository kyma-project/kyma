package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceType string

const (
	SourceTypeGit SourceType = "git"
)

// Runtime enumerates runtimes that are currently supported by Function Controller
// It is a subset of RuntimeExtended
// +kubebuilder:validation:Enum=nodejs12;nodejs14;python38
type Runtime string

const (
	Nodejs12 Runtime = "nodejs12"
	Nodejs14 Runtime = "nodejs14"
	Python38 Runtime = "python38"
)

// RuntimeExtended enumerates runtimes that are either currently supported or
// no longer supported but there still might be "read-only" Functions using them
// +kubebuilder:validation:Enum=nodejs12;nodejs14;python38;nodejs10
type RuntimeExtended string

const (
	RuntimeExtendedNodejs10 RuntimeExtended = "nodejs10"
	RuntimeExtendedNodejs12 RuntimeExtended = "nodejs12"
	RuntimeExtendedNodejs14 RuntimeExtended = "nodejs14"
	RuntimeExtendedPython38 RuntimeExtended = "python38"
)

const (
	ReplicasPresetLabel          = "serverless.kyma-project.io/replicas-preset"
	FunctionResourcesPresetLabel = "serverless.kyma-project.io/function-resources-preset"
	BuildResourcesPresetLabel    = "serverless.kyma-project.io/build-resources-preset"
)

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// Source defines the source code of a function
	Source string `json:"source"`

	// Deps defines the dependencies for a function
	Deps string `json:"deps,omitempty"`

	// +optional
	Runtime Runtime `json:"runtime,omitempty"`

	// Env defines an array of key value pairs need to be used as env variable for a function
	Env []corev1.EnvVar `json:"env,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	BuildResources corev1.ResourceRequirements `json:"buildResources,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`

	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// +optional
	Type SourceType `json:"type,omitempty"`

	Repository `json:",inline,omitempty"`
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

// ConditionType defines condition of function.
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
)

type Condition struct {
	Type               ConditionType          `json:"type,omitempty"`
	Status             corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason        `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	Conditions []Condition `json:"conditions,omitempty"`
	Repository `json:",inline,omitempty"`
	Commit     string          `json:"commit,omitempty"`
	Source     string          `json:"source,omitempty"`
	Runtime    RuntimeExtended `json:"runtime,omitempty"`
}

type Repository struct {
	BaseDir   string `json:"baseDir,omitempty"`
	Reference string `json:"reference,omitempty"`
}

// Function is the Schema for the functions API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='ConfigurationReady')].status"
// +kubebuilder:printcolumn:name="Built",type="string",JSONPath=".status.conditions[?(@.type=='BuildReady')].status"
// +kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
// +kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=".status.runtime"
// +kubebuilder:printcolumn:name="Version",type="integer",JSONPath=".metadata.generation"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

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
