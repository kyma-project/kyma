package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CertificateRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CertificateRequestSpec   `json:"spec"`
	Status            CertificateRequestStatus `json:"status,omitempty"`
}

type CertificateRequestSpec struct {
	CSRInfoURL string `json:"csrInfoUrl"`
}

type CertificateRequestStatus struct {
	Error string `json:"error"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CertificateRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CertificateRequest `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CentralConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CentralConnectionSpec   `json:"spec"`
	Status            CentralConnectionStatus `json:"status,omitempty"`
}

type CentralConnectionSpec struct {
	ManagementInfoURL string      `json:"managementInfoUrl"`
	EstablishedAt     metav1.Time `json:"establishedAt"`
	RenewNow          bool        `json:"renewNow,omitempty"`
}

type CentralConnectionStatus struct {
	SynchronizationStatus SynchronizationStatus   `json:"synchronizationStatus"`
	CertificateStatus     *CertificateStatus      `json:"certificateStatus,omitempty"`
	Error                 *CentralConnectionError `json:"error,omitempty"`
}

type SynchronizationStatus struct {
	LastSync    metav1.Time `json:"lastSync"`
	LastSuccess metav1.Time `json:"lastSuccess"`
}

type CertificateStatus struct {
	NotBefore metav1.Time `json:"notBefore"`
	NotAfter  metav1.Time `json:"notAfter"`
}

type CentralConnectionError struct {
	Message string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CentralConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CentralConnection `json:"items"`
}

func (cc CentralConnection) HasErrorStatus() bool {
	return cc.Status.Error != nil && cc.Status.Error.Message != ""
}

func (cc CentralConnection) HasCertStatus() bool {
	return cc.Status.CertificateStatus != nil && !cc.Status.CertificateStatus.NotAfter.IsZero() && !cc.Status.CertificateStatus.NotBefore.IsZero()
}
