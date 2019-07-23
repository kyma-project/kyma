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
	ManagementInfo ManagementInfo `json:"managementInfo"`
}

type ManagementInfo struct {
	DirectorURL string `json:"directorUrl"`
}

type CompassConnectionStatus struct {
	State                 ConnectionState        `json:"connectionState"`
	ConnectionStatus      *ConnectionStatus      `json:"connectionStatus"`
	SynchronizationStatus *SynchronizationStatus `json:"synchronizationStatus"`
}

func (in CompassConnection) ShouldReconnect() bool {
	return in.Status.State == ConnectionFailed
}

func (s CompassConnectionStatus) String() string {
	//	connectionState := State(s.State)
	//
	//	var certificateError string
	//	if s.CertificateStatus != nil {
	//		certificateError = s.CertificateStatus.SynchronizationFailed
	//	}
	//
	//	var connectionError string
	//	if s.ConnectionStatus != nil {
	//		connectionError = s.ConnectionStatus.SynchronizationFailed
	//	}
	//
	//	var configurationError string
	//	if s.SynchronizationStatus != nil {
	//		configurationError = s.SynchronizationStatus.SynchronizationFailed
	//	}
	//
	//	if certificateError != "" {
	//		certificateError = fmt.Sprintf("Certificate error: %s \n", certificateError)
	//	}
	//
	//	if connectionError != "" {
	//		connectionError = fmt.Sprintf("Connection error: %s \n", connectionError)
	//	}
	//
	//	if configurationError != "" {
	//		configurationError = fmt.Sprintf("Configuration error: %s \n", configurationError)
	//	}
	//
	return ""
}

type ConnectionState string

const (
	// Connection was established successfully
	Connected ConnectionState = "Connected"
	// Connection process failed during authentication to Compass
	ConnectionFailed ConnectionState = "ConnectionFailed"
	// Connection was established but configuration fetching failed
	SynchronizationFailed ConnectionState = "SynchronizationFailed"
	// Connection was established but applying configuration failed
	ResourceApplicationFailed ConnectionState = "ResourceApplicationFailed"
	// Connection was successful and configuration has been applied
	Synchronized ConnectionState = "Synchronized"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CompassConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CompassConnection `json:"items"`
}

// ConnectionStatus represents status of a connection to Compass
type ConnectionStatus struct {
	Established       metav1.Time       `json:"established"`
	LastSync          metav1.Time       `json:"lastSync"`
	LastSuccess       metav1.Time       `json:"lastSuccess"`
	CertificateStatus CertificateStatus `json:"certificateStatus"`
	Error             string            `json:"error,omitempty"`
}

// CertificateStatus represents the status of the certificate
type CertificateStatus struct {
	Acquired  metav1.Time `json:"acquired"`
	NotBefore metav1.Time `json:"notBefore"`
	NotAfter  metav1.Time `json:"notAfter"`
}

// SynchronizationStatus represent the status of Applications synchronization with Compass
type SynchronizationStatus struct {
	LastAttempt               metav1.Time `json:"lastAttempt"`
	LastSuccessfulFetch       metav1.Time `json:"lastSuccessfulFetch"`
	LastSuccessfulApplication metav1.Time `json:"lastSuccessfulApplication"`
	Error                     string      `json:"error,omitempty"`
}
