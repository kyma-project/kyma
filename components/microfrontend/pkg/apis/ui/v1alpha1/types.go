package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MicroFrontend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              MicroFrontendSpec `json:"spec"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterMicroFrontend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterMicroFrontendSpec `json:"spec"`
}

func (rem *MicroFrontend) GetObjectKind() schema.ObjectKind {
	return &MicroFrontend{}
}

func (rem *ClusterMicroFrontend) GetObjectKind() schema.ObjectKind {
	return &ClusterMicroFrontend{}
}

type CommonMicroFrontendSpec struct {
	Version         string           `json:"version"`
	Category        string           `json:"category"`
	ViewBaseURL     string           `json:"viewBaseUrl"`
	NavigationNodes []NavigationNode `json:"navigationNodes"`
}

type MicroFrontendSpec struct {
	CommonMicroFrontendSpec `json:",inline"`
}

type ClusterMicroFrontendSpec struct {
	Placement               string `json:"placement"`
	CommonMicroFrontendSpec `json:",inline"`
}

type NavigationNode struct {
	Label            string `json:"label"`
	NavigationPath   string `json:"navigationPath"`
	ViewURL          string `json:"viewUrl"`
	ShowInNavigation bool   `json:"showInNavigation"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MicroFrontendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MicroFrontend `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterMicroFrontendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterMicroFrontend `json:"items"`
}
