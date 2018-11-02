package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EnvironmentMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
}

func (rem *EnvironmentMapping) GetObjectKind() schema.ObjectKind {
	return &EnvironmentMapping{}
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RemoteEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              RemoteEnvironmentSpec `json:"spec"`
	Status            REStatus              `json:"status,omitempty"`
}

type REStatus struct {
	// Represents the status of Remote Environment release installation
	InstallationStatus InstallationStatus `json:"installationStatus"`
	Conditions         []ReCondition      `json:"conditions"`
}

type InstallationStatus struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

type ReCondition struct {
	// Type of the condition, currently ('Ready').
	Type ReConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string `json:"reason"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string `json:"message"`
}

// ReConditionType represents an Issuer condition value.
type ReConditionType string

const (
	// IssuerConditionReady represents the fact that a given Issuer condition
	// is in ready state.
	Stage1Done ReConditionType = "Stage_1"
	Stage2Done ReConditionType = "Stage_2"
	Stage3Done ReConditionType = "Stage_3"
)

// ConditionStatus represents a condition's status.
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

func (pw *RemoteEnvironment) GetObjectKind() schema.ObjectKind {
	return &RemoteEnvironment{}
}

// RemoteEnvironmentSpec defines spec section of the RemoteEnvironment custom resource
type RemoteEnvironmentSpec struct {
	Description string    `json:"description"`
	Services    []Service `json:"services"`
	// AccessLabel is not required, 'omitempty' is needed because of regexp validation
	AccessLabel string            `json:"accessLabel,omitempty"`
	Labels      map[string]string `json:"labels"`
}

// Entry defines, what is enabled by activating the service.
type Entry struct {
	Type       string `json:"type"`
	GatewayUrl string `json:"gatewayUrl"`
	// AccessLabel is not required for Events, 'omitempty' is needed because of regexp validation
	AccessLabel string      `json:"accessLabel,omitempty"`
	TargetUrl   string      `json:"targetUrl"`
	Credentials Credentials `json:"credentials,omitempty"`
}

// Credentials defines type of authentication and where the credentials are stored
type Credentials struct {
	Type              string `json:"type"`
	SecretName        string `json:"secretName"`
	AuthenticationUrl string `json:"authenticationUrl,omitempty"`
}

// Service represents part of the remote environment, which is mapped 1 to 1 to service class in the service-catalog
type Service struct {
	ID                  string            `json:"id"`
	Identifier          string            `json:"identifier"`
	Name                string            `json:"name"`
	DisplayName         string            `json:"displayName"`
	Description         string            `json:"description"`
	Labels              map[string]string `json:"labels,omitempty"`
	LongDescription     string            `json:"longDescription,omitempty"`
	ProviderDisplayName string            `json:"providerDisplayName"`
	Tags                []string          `json:"tags,omitempty"`
	Entries             []Entry           `json:"entries"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EventActivation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec EventActivationSpec `json:"spec"`
}

type EventActivationSpec struct {
	DisplayName string `json:"displayName"`
	SourceID    string `json:"sourceId"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RemoteEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RemoteEnvironment `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EnvironmentMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EnvironmentMapping `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EventActivationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EventActivation `json:"items"`
}
