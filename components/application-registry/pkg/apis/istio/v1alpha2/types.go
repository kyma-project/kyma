package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rule defines Istio Rule
type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *RuleSpec `json:"spec"`
}

// RuleSpec defines specification for Rule
type RuleSpec struct {
	Match   string       `json:"match"`
	Actions []RuleAction `json:"actions"`
}

// RuleAction defines action for Rule
type RuleAction struct {
	Handler   string   `json:"handler"`
	Instances []string `json:"instances"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleList is a list of Rules
type RuleList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []Rule `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Denier defines Istio Denier
type Denier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *DenierSpec `json:"spec"`
}

// DenierSpec defines specification for Denier
type DenierSpec struct {
	Status *DenierStatus `json:"status"`
}

// DenierStatus defines status for Denier
type DenierStatus struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DenierList is a list of Deniers
type DenierList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []Denier `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Checknothing defines Istio CheckNothing
type Checknothing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ChecknothingList is a list of CheckNothing
type ChecknothingList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []Checknothing `json:"items"`
}
