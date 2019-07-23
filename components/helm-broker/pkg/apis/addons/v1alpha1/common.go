package v1alpha1

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FinalizerAddonsConfiguration defines the finalizer used by Controller, must be qualified name.
const FinalizerAddonsConfiguration string = "addons.kyma-project.io"

// AddonsConfigurationPhase defines the addons configuration phase
type AddonsConfigurationPhase string

const (
	// AddonsConfigurationReady means that Configuration was processed successfully
	AddonsConfigurationReady AddonsConfigurationPhase = "Ready"
	// AddonsConfigurationPending means that Configuration was not yet processed
	AddonsConfigurationPending AddonsConfigurationPhase = "Pending"
	// AddonsConfigurationFailed means that Configuration has some errors
	AddonsConfigurationFailed AddonsConfigurationPhase = "Failed"
)

// AddonStatus define the addon status
type AddonStatus string

const (
	// AddonStatusReady means that given addon is correct
	AddonStatusReady AddonStatus = "Ready"
	// AddonStatusFailed means that there is some problem with the given addon
	AddonStatusFailed AddonStatus = "Failed"
)

// RepositoryStatus define the repository status
type RepositoryStatus string

const (
	// RepositoryStatusFailed means that there is some problem with the given repository
	RepositoryStatusFailed RepositoryStatus = "Failed"

	// RepositoryStatusFailed means that given repository is correct
	RepositoryStatusReady RepositoryStatus = "Ready"
)

// SpecRepository define the addon repository
type SpecRepository struct {
	URL string `json:"url"`
}

// VerifyURL verifies the correctness of the url and completes the url if last parameter is not specified
func (sr *SpecRepository) VerifyURL(developMode bool) error {
	u, err := url.ParseRequestURI(sr.URL)
	if err != nil {
		return errors.Wrap(err, "while parsing URL")
	}
	repositoryURL := u.String()

	// check if yaml file at the end of url exists if not add default 'index.yaml'
	extension := filepath.Ext(path.Base(repositoryURL))
	if extension != ".yaml" {
		repositoryURL = strings.TrimRight(repositoryURL, "/") + "/index.yaml"
	}

	// check working mode if production, check security of url
	if developMode {
		sr.URL = repositoryURL
		return nil
	}

	if u.Scheme != "https" {
		return fmt.Errorf("Repository URL %s is unsecured", repositoryURL)
	}

	sr.URL = repositoryURL
	return nil
}

// CommonAddonsConfigurationSpec defines the desired state of (Cluster)AddonsConfiguration
type CommonAddonsConfigurationSpec struct {
	// ReprocessRequest is strictly increasing, non-negative integer counter
	// that can be incremented by a user to manually trigger the reprocessing action of given CR.
	ReprocessRequest uint64           `json:"reprocessRequest,omitempty"`
	Repositories     []SpecRepository `json:"repositories"`
}

// Addon holds information about single addon
type Addon struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// +kubebuilder:validation:Enum=Ready,Failed
	Status  AddonStatus       `json:"status,omitempty"`
	Reason  AddonStatusReason `json:"reason,omitempty"`
	Message string            `json:"message,omitempty"`
}

// Key returns a key for an addon
func (a *Addon) Key() string {
	return a.Name + "/" + a.Version
}

// StatusRepository define the addon repository
type StatusRepository struct {
	URL     string                 `json:"url"`
	Status  RepositoryStatus       `json:"status,omitempty"`
	Reason  RepositoryStatusReason `json:"reason,omitempty"`
	Message string                 `json:"message,omitempty"`
	Addons  []Addon                `json:"addons"`
}

// CommonAddonsConfigurationStatus defines the observed state of AddonsConfiguration
type CommonAddonsConfigurationStatus struct {
	Phase              AddonsConfigurationPhase `json:"phase"`
	LastProcessedTime  *metav1.Time             `json:"lastProcessedTime,omitempty"`
	ObservedGeneration int64                    `json:"observedGeneration,omitempty"`
	Repositories       []StatusRepository       `json:"repositories,omitempty"`
}
