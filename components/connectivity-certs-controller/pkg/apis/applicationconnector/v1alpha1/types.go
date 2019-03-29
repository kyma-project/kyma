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

// TODO - think of better name
type CentralConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CentralConnectionSpec   `json:"spec"`
	Status            CentralConnectionStatus `json:"status,omitempty"`
}

type CentralConnectionSpec struct {
	ManagementInfoURL string `json:"managementInfoUrl"`
}

type CentralConnectionStatus struct {
	LastSync           int64                   `json:"lastSync"`
	CertificateValidTo int64                   `json:"certificateValidTo"`
	Error              *CentralConnectionError `json:"error,omitempty"`
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
