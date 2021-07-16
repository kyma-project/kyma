package v1alpha1

import (
	"crypto/x509"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster
type CompassConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CompassConnectionSpec   `json:"spec"`
	Status            CompassConnectionStatus `json:"status,omitempty"`
}

type CompassConnectionSpec struct {
	ManagementInfo        ManagementInfo `json:"managementInfo"`
	ResyncNow             bool           `json:"resyncNow,omitempty"`
	RefreshCredentialsNow bool           `json:"refreshCredentialsNow,omitempty"`
}

type ManagementInfo struct {
	DirectorURL  string `json:"directorUrl"`
	ConnectorURL string `json:"connectorUrl"`
}

type CompassConnectionStatus struct {
	State            ConnectionState   `json:"connectionState"`
	ConnectionStatus *ConnectionStatus `json:"connectionStatus"`
	// +optional
	// +nullable
	SynchronizationStatus *SynchronizationStatus `json:"synchronizationStatus"`
}

func (in *CompassConnection) SetCertificateStatus(acquired metav1.Time, certificate *x509.Certificate) {
	if in.Status.ConnectionStatus == nil {
		in.Status.ConnectionStatus = &ConnectionStatus{}
	}

	in.Status.ConnectionStatus.CertificateStatus = CertificateStatus{
		Acquired:  acquired,
		NotBefore: metav1.NewTime(certificate.NotBefore),
		NotAfter:  metav1.NewTime(certificate.NotAfter),
	}
}

func (in CompassConnection) ShouldAttemptReconnect() bool {
	return in.Status.State == ConnectionFailed
}

func (in CompassConnection) ShouldRenewCertificate(certValidityRenewalThreshold float64, minimalSyncTime time.Duration) bool {
	if in.Spec.RefreshCredentialsNow {
		return true
	}

	notBefore := in.Status.ConnectionStatus.CertificateStatus.NotBefore.Unix()
	notAfter := in.Status.ConnectionStatus.CertificateStatus.NotAfter.Unix()

	certValidity := notAfter - notBefore

	timeLeft := float64(notAfter - time.Now().Unix())

	return timeLeft < float64(certValidity)*certValidityRenewalThreshold || timeLeft < 2*minimalSyncTime.Seconds()
}

func (s CompassConnectionStatus) String() string {
	// TODO: return more detailed status
	return string(s.State)
}

type ConnectionProcessStatus struct {
	ConnectionEstablished bool
}

type ConnectionState string

// TODO - consider reworking those states to some pipeline info

const (
	// Connection was established successfully
	Connected ConnectionState = "Connected"
	// Connection process failed during authentication to Compass
	ConnectionFailed ConnectionState = "ConnectionFailed"
	// Connection was established but the error occurred during connection maintenance
	ConnectionMaintenanceFailed ConnectionState = "ConnectionMaintenanceFailed"
	// Connection was established but configuration fetching failed
	SynchronizationFailed ConnectionState = "SynchronizationFailed"
	// Connection was established but applying configuration failed
	ResourceApplicationFailed ConnectionState = "ResourceApplicationFailed"
	// Resources were applied successfully but Runtime metadata update failed
	MetadataUpdateFailed ConnectionState = "MetadataUpdateFailed"
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
	// +optional
	// +nullable
	Established metav1.Time `json:"established"`
	// +optional
	// +nullable
	Renewed metav1.Time `json:"renewed,omitempty"`
	// +optional
	// +nullable
	LastSync metav1.Time `json:"lastSync"`
	// +optional
	// +nullable
	LastSuccess       metav1.Time       `json:"lastSuccess"`
	CertificateStatus CertificateStatus `json:"certificateStatus"`
	Error             string            `json:"error,omitempty"`
}

// CertificateStatus represents the status of the certificate
type CertificateStatus struct {
	// +optional
	// +nullable
	Acquired metav1.Time `json:"acquired"`
	// +optional
	// +nullable
	NotBefore metav1.Time `json:"notBefore"`
	// +optional
	// +nullable
	NotAfter metav1.Time `json:"notAfter"`
}

// SynchronizationStatus represent the status of Applications synchronization with Compass
type SynchronizationStatus struct {
	// +optional
	// +nullable
	LastAttempt metav1.Time `json:"lastAttempt"`
	// +optional
	// +nullable
	LastSuccessfulFetch metav1.Time `json:"lastSuccessfulFetch"`
	// +optional
	// +nullable
	LastSuccessfulApplication metav1.Time `json:"lastSuccessfulApplication"`
	Error                     string      `json:"error,omitempty"`
}
