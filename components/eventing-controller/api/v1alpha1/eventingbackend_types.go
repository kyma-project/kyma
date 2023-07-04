package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=BEB;NATS

type BackendType string

const (
	BEBBackendType  BackendType = "BEB"
	NatsBackendType BackendType = "NATS"
)

// EventingBackendSpec defines the desired state of EventingBackend.
type EventingBackendSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// EventingBackendStatus defines the observed state of EventingBackend.
type EventingBackendStatus struct {
	// Specifies the backend type used. The value is either `BEB`, or `NATS`.
	// +optional
	Backend BackendType `json:"backendType"`

	// Defines the overall Backend status.
	// +optional
	EventingReady *bool `json:"eventingReady"`

	// Defines the status of the Controller and the EPP.
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// Name of the Secret containing BEB access tokens, required for BEB only.
	// +optional
	BEBSecretName string `json:"bebSecretName,omitempty"`

	// Namespace of the Secret containing BEB access tokens, required for BEB only.
	// +optional
	BEBSecretNamespace string `json:"bebSecretNamespace,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Backend",type=string,JSONPath=`.status.backendType`
// +kubebuilder:printcolumn:name="EventingReady",type=boolean,JSONPath=`.status.eventingReady`
// +kubebuilder:printcolumn:name="SubscriptionControllerReady",type=string,JSONPath=`.status.conditions[?(@.type=="Subscription Controller Ready")].status`
// +kubebuilder:printcolumn:name="PublisherProxyReady",type=string,JSONPath=`.status.conditions[?(@.type=="Publisher Proxy Ready")].status`

// EventingBackend is the Schema for the eventingbackends API.
type EventingBackend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventingBackendSpec   `json:"spec,omitempty"`
	Status EventingBackendStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EventingBackendList contains a list of EventingBackend.
type EventingBackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventingBackend `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&EventingBackend{}, &EventingBackendList{})
}
