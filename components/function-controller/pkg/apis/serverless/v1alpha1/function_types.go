package v1alpha1

import (
	"crypto/sha256"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceType string

const (
	SourceTypeGit SourceType = "git"
)

// Runtime enumerates runtimes that are currently supported by Function Controller
// It is a subset of RuntimeExtended
// +kubebuilder:validation:Enum=nodejs12;nodejs14;python39
type Runtime string

const (
	Nodejs12 Runtime = "nodejs12"
	Nodejs14 Runtime = "nodejs14"
	Python39 Runtime = "python39"
)

// RuntimeExtended enumerates runtimes that are either currently supported or
// no longer supported but there still might be "read-only" Functions using them
// +kubebuilder:validation:Enum=nodejs12;nodejs14;nodejs10;python38;python39
type RuntimeExtended string

const (
	RuntimeExtendedNodejs10 RuntimeExtended = "nodejs10"
	RuntimeExtendedNodejs12 RuntimeExtended = "nodejs12"
	RuntimeExtendedNodejs14 RuntimeExtended = "nodejs14"
	RuntimeExtendedPython38 RuntimeExtended = "python38"
	RuntimeExtendedPython39 RuntimeExtended = "python39"
)

const (
	ReplicasPresetLabel          = "serverless.kyma-project.io/replicas-preset"
	FunctionResourcesPresetLabel = "serverless.kyma-project.io/function-resources-preset"
	BuildResourcesPresetLabel    = "serverless.kyma-project.io/build-resources-preset"
)

const (
	FunctionSourceKey = "source"
	FunctionDepsKey   = "dependencies"
)

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// Source defines the source code of a function
	Source string `json:"source"`

	// Deps defines the dependencies for a function
	Deps string `json:"deps,omitempty"`

	// +optional
	Runtime Runtime `json:"runtime,omitempty"`

	// +optional
	RuntimeImageOverride string `json:"runtimeImageOverride,omitempty"`

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
	ConditionReasonGitAuthorizationFailed         ConditionReason = "GitAuthorizationFailed"
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
	Type               ConditionType          `json:"type,omitempty"`
	Status             corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason        `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

// FunctionStatus defines the observed state of Function
type FunctionStatus struct {
	Conditions           []Condition `json:"conditions,omitempty"`
	Repository           `json:",inline,omitempty"`
	Commit               string          `json:"commit,omitempty"`
	Source               string          `json:"source,omitempty"`
	Runtime              RuntimeExtended `json:"runtime,omitempty"`
	RuntimeImageOverride string          `json:"runtimeImageOverride,omitempty"`
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

func (s *FunctionStatus) Condition(c ConditionType) *Condition {
	for _, cond := range s.Conditions {
		if cond.Type == c {
			return &cond
		}
	}
	return nil
}

func (c *Condition) IsTrue() bool {
	return c.Status == corev1.ConditionTrue
}

func (c *Condition) HasStatus(s corev1.ConditionStatus) bool {
	return c != nil && c.Status == s
}

func (f *Function) GenerateInternalLabels() map[string]string {
	labels := make(map[string]string, 3)

	labels[FunctionNameLabel] = f.Name
	labels[FunctionManagedByLabel] = FunctionControllerValue
	labels[FunctionUUIDLabel] = string(f.GetUID())
	return labels
}

func (f *Function) GetMergedLables() map[string]string {
	internalLabels := f.GenerateInternalLabels()
	functionLabels := f.GetLabels()

	return f.MergeLabels(functionLabels, internalLabels)
}

func (f *Function) MergeLabels(labelsCollection ...map[string]string) map[string]string {
	result := make(map[string]string, 0)
	for _, labels := range labelsCollection {
		for key, value := range labels {
			result[key] = value
		}
	}
	return result
}

func (f *Function) CalculateImageTag() string {
	hash := sha256.Sum256([]byte(strings.Join([]string{
		string(f.GetUID()),
		f.Spec.Source,
		f.Spec.Deps,
		string(f.Status.Runtime),
	}, "-")))

	return fmt.Sprintf("%x", hash)
}

func (f *Function) CalculateGitImageTag() string {
	data := strings.Join([]string{
		string(f.GetUID()),
		f.Status.Commit,
		f.Status.Repository.BaseDir,
		string(f.Status.Runtime),
	}, "-")
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func (f *Function) BuildImageAddress(registryAddress string) string {
	var imageTag string
	if f.Spec.Type == SourceTypeGit {
		imageTag = f.CalculateGitImageTag()
	} else {
		imageTag = f.CalculateImageTag()
	}
	return fmt.Sprintf("%s/%s-%s:%s", registryAddress, f.Namespace, f.Name, imageTag)
}

func (f *Function) GetDefaultReplicas() (int32, int32) {
	min, max := int32(1), int32(1)
	spec := f.Spec
	if spec.MinReplicas != nil && *spec.MinReplicas > 0 {
		min = *spec.MinReplicas
	}
	// special case
	if spec.MaxReplicas == nil || min > *spec.MaxReplicas {
		max = min
	} else {
		max = *spec.MaxReplicas
	}
	return min, max
}

func (f *Function) gitConfigChanged(commit string) bool {
	changedCommits := f.Status.Commit == "" || commit != f.Status.Commit
	changedReferences := f.Spec.Reference != f.Status.Reference
	changedBaseDirs := f.Spec.BaseDir != f.Status.BaseDir

	return changedCommits || changedReferences || changedBaseDirs
}
func (f *Function) GitSourceChanged(commit string) bool {
	gitConfigChanged := f.gitConfigChanged(commit)
	changedRuntimes := RuntimeExtended(f.Spec.Runtime) != f.Status.Runtime
	notConfigured := f.Status.Condition(ConditionConfigurationReady).HasStatus(corev1.ConditionFalse)

	return gitConfigChanged || changedRuntimes || notConfigured
}

func (f *Function) DeploymentSelectorLabels() map[string]string {
	labels := f.GenerateInternalLabels()
	labels[FunctionResourceLabel] = FunctionResourceLabelDeploymentValue
	return labels
}
