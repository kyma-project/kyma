package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApplicationMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ApplicationMappingSpec `json:"spec"`
}

func (in *ApplicationMapping) GetObjectKind() schema.ObjectKind {
	return &ApplicationMapping{}
}

type ApplicationMappingSpec struct {
	Services []ApplicationMappingService `json:"services"`
}

type ApplicationMappingService struct {
	ID string `json:"id"`
}

// IsAllApplicationServicesEnabled returns true if the mapping enables whole Application.
// It means every existing and new service (added in the future) in the Application is enabled by the ApplicationMapping.
// The method returns true if the Spec.Services list is nil.
func (in *ApplicationMapping) IsAllApplicationServicesEnabled() bool {
	return in.Spec.Services == nil
}

// IsServiceEnabled returns true if the service with given ID is enabled
func (in *ApplicationMapping) IsServiceEnabled(id string) bool {
	if in.IsAllApplicationServicesEnabled() {
		return true
	}
	for _, svc := range in.Spec.Services {
		if svc.ID == id {
			return true
		}
	}
	return false
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EventActivation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec EventActivationSpec `json:"spec"`
}

func (in *EventActivation) GetObjectKind() schema.ObjectKind {
	return &EventActivation{}
}

type EventActivationSpec struct {
	DisplayName string `json:"displayName"`
	SourceID    string `json:"sourceId"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApplicationMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ApplicationMapping `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EventActivationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EventActivation `json:"items"`
}
