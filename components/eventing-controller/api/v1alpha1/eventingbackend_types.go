package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=BEB;NATS
type BackendType string

const (
	BebBackendType  BackendType = "BEB"
	NatsBackendType BackendType = "NATS"
)

// EventingBackendSpec defines the desired state of EventingBackend
type EventingBackendSpec struct {
	// For now spec is empty!
}

// EventingBackendStatus defines the observed state of EventingBackend
type EventingBackendStatus struct {
	// Specifies the backend type used. Allowed values are "BEB" and "NATS"
	// +optional
	Backend BackendType `json:"backendType"`

	// +optional
	EventingReady *bool `json:"eventingReady"`

	// +optional
	SubscriptionControllerReady *bool `json:"subscriptionControllerReady"`

	// +optional
	PublisherProxyReady *bool `json:"publisherProxyReady"`

	// The name of the secret containing BEB access tokens, required only for BEB
	// +optional
	BebSecretName string `json:"bebSecretName"`

	// The namespace of the secret containing BEB access tokens, required only for BEB
	// +optional
	BebSecretNamespace string `json:"bebSecretNamespace"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Backend",type=string,JSONPath=`.status.backendType`
// +kubebuilder:printcolumn:name="EventingReady",type=boolean,JSONPath=`.status.eventingReady`
// +kubebuilder:printcolumn:name="SubscriptionControllerReady",type=boolean,JSONPath=`.status.subscriptionControllerReady`
// +kubebuilder:printcolumn:name="PublisherProxyReady",type=boolean,JSONPath=`.status.publisherProxyReady`
// EventingBackend is the Schema for the eventingbackends API
type EventingBackend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventingBackendSpec   `json:"spec,omitempty"`
	Status EventingBackendStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EventingBackendList contains a list of EventingBackend
type EventingBackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventingBackend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EventingBackend{}, &EventingBackendList{})
}
