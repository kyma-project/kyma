package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&CompassConnection{}, &CompassConnectionList{})
}

// TODO - what should be full CRD domain? compassconnection.compass.kyma-project.io ?

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
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CompassConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CompassConnection `json:"items"`
}

//type CentralConnectionSpec struct {
//	ManagementInfoURL string      `json:"managementInfoUrl"`
//	EstablishedAt     metav1.Time `json:"establishedAt"`
//	RenewNow          bool        `json:"renewNow,omitempty"`
//}
//
//type CentralConnectionStatus struct {
//	SynchronizationStatus SynchronizationStatus   `json:"synchronizationStatus"`
//	CertificateStatus     *CertificateStatus      `json:"certificateStatus,omitempty"`
//	Error                 *CentralConnectionError `json:"error,omitempty"`
//}
//
//type SynchronizationStatus struct {
//	LastSync    metav1.Time `json:"lastSync"`
//	LastSuccess metav1.Time `json:"lastSuccess"`
//}
//
//type CertificateStatus struct {
//	NotBefore metav1.Time `json:"notBefore"`
//	NotAfter  metav1.Time `json:"notAfter"`
//}
//
//type CentralConnectionError struct {
//	Message string `json:"message"`
//}
//
//// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//
//type CentralConnectionList struct {
//	metav1.TypeMeta `json:",inline"`
//	metav1.ListMeta `json:"metadata"`
//
//	Items []CentralConnection `json:"items"`
//}
//
//func (cc CentralConnection) HasErrorStatus() bool {
//	return cc.Status.Error != nil && cc.Status.Error.Message != ""
//}
//
//func (cc CentralConnection) HasCertStatus() bool {
//	return cc.Status.CertificateStatus != nil && !cc.Status.CertificateStatus.NotAfter.IsZero() && !cc.Status.CertificateStatus.NotBefore.IsZero()
//}
