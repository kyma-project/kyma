package v1alpha1

import (
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	BindingLabelKey          = "bindings.kyma.project.io"
	BindingValidatedLabelKey = "validation.bindings.kyma.project.io"
)

const (
	BindingFinalizer string = "kyma-project-io/bindings"
)

type BindingPhase string

const (
	BindingReady   = "Ready"
	BindingPending = "Pending"
	BindingFailed  = "Failed"
)

type SourceKind string

const (
	SourceKindSecret    = "secret"
	SourceKindConfigMap = "config-map"
)

type Source struct {
	Kind SourceKind `json:"kind"`
	Name string     `json:"name"`
}

type Target struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// BindingSpec defines the desired state of Binding
type BindingSpec struct {
	Source Source `json:"source"`
	Target Target `json:"target"`
}

// BindingStatus defines the observed state of Binding
type BindingStatus struct {
	Phase             BindingPhase `json:"phase"`
	Message           string       `json:"message"`
	Target            string       `json:"target"`
	Source            string       `json:"source"`
	LastProcessedTime *metav1.Time `json:"lastProcessedTime,omitempty"`
}

func (bs *BindingStatus) IsEmpty() bool {
	return bs.Phase == ""
}

func (bs *BindingStatus) IsPending() bool {
	return bs.Phase == BindingPending
}

func (bs *BindingStatus) IsFailed() bool {
	return bs.Phase == BindingFailed
}

func (bs *BindingStatus) Init() error {
	if bs.Phase != "" {
		return errors.New("status cannot be initialized from state other than empty")
	}
	bs.Phase = BindingPending
	bs.LastProcessedTime = &metav1.Time{Time: time.Now()}

	return nil
}

func (bs *BindingStatus) Pending() error {
	if bs.Phase == "" {
		return errors.New("status cannot be set to Pending from empty state")
	}
	bs.Phase = BindingPending
	bs.LastProcessedTime = &metav1.Time{Time: time.Now()}

	return nil
}

func (bs *BindingStatus) Ready() error {
	if bs.Phase == "" {
		return errors.New("status cannot be set to Ready from empty state")
	}
	bs.Phase = BindingReady
	bs.LastProcessedTime = &metav1.Time{Time: time.Now()}

	return nil
}

func (bs *BindingStatus) Failed() error {
	if bs.Phase == "" {
		return errors.New("status cannot be set to Failed from empty state")
	}
	bs.Phase = BindingFailed
	bs.LastProcessedTime = &metav1.Time{Time: time.Now()}

	return nil
}

// Binding is the Schema for the bindings API
type Binding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BindingSpec   `json:"spec,omitempty"`
	Status BindingStatus `json:"status,omitempty"`
}

// BindingList contains a list of Binding
type BindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Binding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Binding{}, &BindingList{})
}
