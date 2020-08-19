package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceBindingUsage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceBindingUsageSpec   `json:"spec"`
	Status            ServiceBindingUsageStatus `json:"status"`
}

func (pw *ServiceBindingUsage) GetObjectKind() schema.ObjectKind {
	return &ServiceBindingUsage{}
}

// Status represents the current and past information about binding usage
type ServiceBindingUsageStatus struct {
	// Conditions represents the observations of a ServiceBindingUsage's state.
	Conditions []ServiceBindingUsageCondition `json:"conditions,omitempty"`
}

// ServiceBindingUsageCondition describes the state of a ServiceBindingUsage at a certain point.
type ServiceBindingUsageCondition struct {
	// Type is the type of the condition.
	Type ServiceBindingUsageConditionType `json:"type"`
	// Status is the status of the condition.
	Status ConditionStatus `json:"status"`
	// The last time this condition was updated.
	// LastUpdateTime will be updated even when the status of the condition is not changed
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// ServiceBindingUsageConditionType represents a usage condition value.
type ServiceBindingUsageConditionType string

const (
	// ServiceBindingUsageReady represents the fact that a given usage is in ready state.
	ServiceBindingUsageReady ServiceBindingUsageConditionType = "Ready"
)

// ConditionStatus represents a condition's status.
type ConditionStatus string

const (
	// ConditionTrue means a resource is in the condition.
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse means a resource is not in the condition
	ConditionFalse ConditionStatus = "False"
	// ConditionUnknown means controller can't decide if a usage is in the condition or not.
	ConditionUnknown ConditionStatus = "Unknown"
)

// ServiceBindingUsageSpec represents a description of the ServiceBindingUsage
type ServiceBindingUsageSpec struct {
	// ReprocessRequest is strictly increasing, non-negative integer counter
	// that can be incremented by a user to manually trigger the reprocessing action of given CR.
	ReprocessRequest int64 `json:"reprocessRequest,omitempty"`
	// ServiceBindingRef is the reference to the ServiceBinding and
	// need to be in the same namespace where ServiceBindingUsage was created.
	ServiceBindingRef LocalReferenceByName `json:"serviceBindingRef"`
	// UsedBy is the reference to the application which should be configured to use ServiceInstance pointed by serviceBindingRef.
	// Pointed resource should be available in the same namespace where ServiceBindingUsage was created.
	UsedBy LocalReferenceByKindAndName `json:"usedBy"`
	// Parameters is a set of the parameters passed to the controller
	Parameters *Parameters `json:"parameters,omitempty"`
}

// LocalReferenceByName contains enough information to let you locate the
// referenced object inside the same namespace.
type LocalReferenceByName struct {
	// Name of the referent.
	Name string `json:"name"`
}

// Parameters contain all parameters which are used by controller
type Parameters struct {
	// EnvPrefix defines the prefix for environment variables injected from ServiceBinding
	EnvPrefix *EnvPrefix `json:"envPrefix,omitempty"`
}

// EnvPrefix defines the prefixing of environment variables
type EnvPrefix struct {
	// Name of the prefix
	Name string `json:"name"`
}

// LocalReferenceByKindAndName contains enough information to let you locate the
// referenced to generic object inside the same namespace.
type LocalReferenceByKindAndName struct {
	// Name of the referent
	Name string `json:"name"`
	// Kind of the referent
	Kind string `json:"kind"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceBindingUsageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceBindingUsage `json:"items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type UsageKind struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              UsageKindSpec `json:"spec"`
}

// UsageKindSpec represents a description of the ServiceBindingTarget
type UsageKindSpec struct {
	DisplayName string             `json:"displayName"`
	Resource    *ResourceReference `json:"resource"`
	LabelsPath  string             `json:"labelsPath"`
}

type ResourceReference struct {
	Group   string `json:"group"`
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type UsageKindList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []UsageKind `json:"items"`
}
