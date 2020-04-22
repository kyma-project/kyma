package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FunctionSpec defines the desired state of Function
type FunctionSpec struct {
	// Source defines the source code of a function
	Source string `json:"source"`

	// Deps defines the dependencies for a function
	Deps string `json:"deps,omitempty"`

	// Env defines an array of key value pairs need to be used as env variable for a function
	Env []corev1.EnvVar `json:"env,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Minimum:=0
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Minimum:=0
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
}

const (
	FunctionNameLabel      = "serverless.kyma-project.io/function-name"
	FunctionManagedByLabel = "serverless.kyma-project.io/managed-by"
	FunctionUUIDLabel      = "serverless.kyma-project.io/uuid"
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
	ConditionReasonConfigMapCreated ConditionReason = "ConfigMapCreated"
	ConditionReasonConfigMapUpdated ConditionReason = "ConfigMapUpdated"
	ConditionReasonConfigMapError   ConditionReason = "ConfigMapError"
	ConditionReasonJobFailed        ConditionReason = "JobFailed"
	ConditionReasonJobCreated       ConditionReason = "JobCreated"
	ConditionReasonJobRunning       ConditionReason = "JobRunning"
	ConditionReasonJobsDeleted      ConditionReason = "JobsDeleted"
	ConditionReasonJobFinished      ConditionReason = "JobFinished"
	ConditionReasonServiceCreated   ConditionReason = "ServiceCreated"
	ConditionReasonServiceUpdated   ConditionReason = "ServiceUpdated"
	ConditionReasonServiceFailed    ConditionReason = "ServiceFailed"
	ConditionReasonServiceWaiting   ConditionReason = "ServiceWaiting"
	ConditionReasonServiceReady     ConditionReason = "ServiceReady"
)

type Condition struct {
	Type               ConditionType          `json:"type,omitempty"`
	Status             corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason        `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

// FunctionStatus defines the observed state of FuncSONPath: .status.phase
type FunctionStatus struct {
	Conditions []Condition `json:"conditions,omitempty"`
}

// Function is the Schema for the functions API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Configured",type="string",JSONPath=".status.conditions[?(@.type=='ConfigurationReady')].status"
// +kubebuilder:printcolumn:name="Built",type="string",JSONPath=".status.conditions[?(@.type=='BuildReady')].status"
// +kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
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
