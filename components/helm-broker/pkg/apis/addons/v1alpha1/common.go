package v1alpha1

import (
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
