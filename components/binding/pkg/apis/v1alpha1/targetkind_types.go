package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kind represents Kubernetes Kind name
type Kind string

// TargetKindPhase describes TargetKind phase
type TargetKindPhase string

const (
	TargetKindRegistered = "Registered"
	TargetKindFailed     = "Failed"
)

// TargetKindSpec defines the desired state of TargetKind
type TargetKindSpec struct {
	DisplayName string   `json:"displayName"`
	Resource    Resource `json:"resource"`
	LabelsPath  string   `json:"labelsPath"`
}

type Resource struct {
	Group   string `json:"group"`
	Kind    Kind   `json:"kind"`
	Version string `json:"version"`
}

// TargetKindStatus defines the observed state of TargetKind
type TargetKindStatus struct {
	Phase             TargetKindPhase `json:"phase"`
	Message           string          `json:"message"`
	LastProcessedTime *metav1.Time    `json:"lastProcessedTime,omitempty"`
}

func (tks *TargetKindStatus) IsEmpty() bool {
	return tks.Phase == ""
}

func (tks *TargetKindStatus) IsRegistered() bool {
	return tks.Phase == TargetKindRegistered
}

func (tks *TargetKindStatus) Registered() error {
	tks.Phase = TargetKindRegistered
	tks.LastProcessedTime = &metav1.Time{Time: time.Now()}

	return nil
}

func (tks *TargetKindStatus) IsFailed() bool {
	return tks.Phase == TargetKindFailed
}

func (tks *TargetKindStatus) Failed() error {
	tks.Phase = TargetKindFailed
	tks.LastProcessedTime = &metav1.Time{Time: time.Now()}

	return nil
}

// +kubebuilder:object:root=true

// TargetKind is the Schema for the targetkinds API
type TargetKind struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetKindSpec   `json:"spec,omitempty"`
	Status TargetKindStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TargetKindList contains a list of TargetKind
type TargetKindList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetKind `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TargetKind{}, &TargetKindList{})
}
