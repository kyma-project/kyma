package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitRepositorySpec defines the desired state of GitRepository
type GitRepositorySpec struct {

	// +kubebuilder:validation:Required

	// URL is the address of GIT repository
	URL string `json:"url"`

	// Auth is the optional definition of authentication that should be used for repository operations
	// +optional
	Auth *RepositoryAuth `json:"auth,omitempty"`
}

// RepositoryAuth defines authentication method used for repository operations
type RepositoryAuth struct {
	// Type is the type of authentication
	Type RepositoryAuthType `json:"type"`

	// +kubebuilder:validation:Required

	// SecretName is the name of Kubernetes Secret containing credentials used for authentication
	SecretName string `json:"secretName"`
}

// RepositoryAuthType is the enum of available authentication types
// +kubebuilder:validation:Enum=basic;key
type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey                    = "key"
)

// +kubebuilder:object:root=true

// GitRepository is the Schema for the gitrepositories API
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Auth",type=string,JSONPath=`.spec.auth.type`
type GitRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GitRepositorySpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// GitRepositoryList contains a list of GitRepository
type GitRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitRepository{}, &GitRepositoryList{})
}
