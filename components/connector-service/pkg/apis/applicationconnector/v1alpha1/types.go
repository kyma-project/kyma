package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KymaGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              KymaGroupSpec   `json:"spec"`
	Status            KymaGroupStatus `json:"status,omitempty"`
}

type KymaGroupStatus struct {
}

func (pw *KymaGroup) GetObjectKind() schema.ObjectKind {
	return &KymaGroup{}
}

// KymaGroupSpec defines spec section of the KymaGroup custom resource
type KymaGroupSpec struct {
	DisplayName  string        `json:"displayName"`
	Tenant       string        `json:"tenant,required"`
	Cluster      Cluster       `json:"cluster,omitempty"`
	Applications []Application `json:"applications"`
}

// Cluster defines basic information about the cluster
type Cluster struct {
	AppRegistryUrl string `json:"appRegistryUrl"`
	EventsUrl      string `json:"eventsUrl"`
}

// Application represents the basic information about the Application in the group
type Application struct {
	ID string `json:"id"`
	// TODO - display name will be needed for validation unique combination
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KymaGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KymaGroup `json:"items"`
}
