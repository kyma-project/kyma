package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ApplicationSpec   `json:"spec"`
	Status            ApplicationStatus `json:"status,omitempty"`
}

type ApplicationStatus struct {
	// Represents the status of Application release installation
	InstallationStatus InstallationStatus `json:"installationStatus"`
}

type InstallationStatus struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func (pw *Application) GetObjectKind() schema.ObjectKind {
	return &Application{}
}

// ApplicationSpec defines spec section of the Application custom resource
type ApplicationSpec struct {
	Description      string    `json:"description"`
	SkipInstallation bool      `json:"skipInstallation,omitempty"`
	Services         []Service `json:"services"`
	// AccessLabel is not required, 'omitempty' is needed because of regexp validation
	AccessLabel string            `json:"accessLabel,omitempty"`
	Labels      map[string]string `json:"labels"`
	Tenant      string            `json:"tenant,omitempty"`
	Group       string            `json:"group,omitempty"`
}

// Entry defines, what is enabled by activating the service.
type Entry struct {
	Type       string `json:"type"`
	GatewayUrl string `json:"gatewayUrl"`
	// AccessLabel is not required for Events, 'omitempty' is needed because of regexp validation
	AccessLabel      string      `json:"accessLabel,omitempty"`
	TargetUrl        string      `json:"targetUrl"`
	SpecificationUrl string      `json:"specificationUrl,omitempty"`
	ApiType          string      `json:"apiType,omitempty"`
	Credentials      Credentials `json:"credentials,omitempty"`
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Application `json:"items"`
}
