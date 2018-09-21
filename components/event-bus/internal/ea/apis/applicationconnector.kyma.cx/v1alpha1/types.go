package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventActivation describes an EventActivation.
type EventActivation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	EventActivationSpec `json:"spec"`
}

// EventActivationSpec for Event Activation of EventBus
type EventActivationSpec struct {
	DisplayName string `json:"displayName"`
	SourceID    string `json:"sourceId"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventActivationList is a list of EventActivation resources
type EventActivationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EventActivation `json:"items"`
}
