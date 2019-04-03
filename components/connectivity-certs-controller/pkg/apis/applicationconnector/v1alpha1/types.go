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
	EstablishedAt     metav1.Time `json:"establishedAt"` // TODO - should it be a part of Spec or Status
	RenewNow          bool        `json:"renewNow"`
}

type CentralConnectionStatus struct {
	SynchronizationStatus SynchronizationStatus   `json:"synchronizationStatus"`
	CertificateStatus     CertificateStatus       `json:"certificateStatus"`
	Error                 *CentralConnectionError `json:"error,omitempty"`
}

type SynchronizationStatus struct {
	LastSync    metav1.Time `json:"lastSync"`
	LastSuccess metav1.Time `json:"lastSuccess"`
}

type CertificateStatus struct {
	IssuedAt metav1.Time `json:"issuedAt"`
	ValidTo  metav1.Time `json:"validTo"`
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
