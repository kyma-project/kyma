package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CompassConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CompassConnectionSpec   `json:"spec"`
	Status            CompassConnectionStatus `json:"status,omitempty"`
}

type CompassConnectionSpec struct {
}

type CompassConnectionStatus struct {
	ConnectionState ConnectionState `json:"connectionState"`
}

type ConnectionState string

const (
	NotConnected ConnectionState = "NotConnected"
	Connected    ConnectionState = "Connected"
	Error        ConnectionState = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CompassConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CompassConnection `json:"items"`
}

type SynchronizationStatus struct {
	LastSync    metav1.Time `json:"lastSync"`
	LastSuccess metav1.Time `json:"lastSuccess"`
	Error       string      `json:"error,omitempty"`
}

type CertificateStatus struct {
	NotBefore metav1.Time `json:"notBefore"`
	NotAfter  metav1.Time `json:"notAfter"`
	Error     string      `json:"error,omitempty"`
}
