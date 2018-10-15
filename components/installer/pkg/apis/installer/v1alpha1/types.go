package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Installation .
type Installation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstallationSpec   `json:"spec"`
	Status InstallationStatus `json:"status"`
}

// ShouldInstall returns true when user requested install action
func (i *Installation) ShouldInstall() bool {
	action := i.ObjectMeta.Labels["action"]

	return action == ActionInstall && i.canInstall()
}

func (i *Installation) canInstall() bool {
	return (i.Status.State == StateEmpty || i.Status.State == StateUninstalled || i.Status.State == StateInstalled)
}

// ShouldUninstall returns true when user requested uninstall action
func (i *Installation) ShouldUninstall() bool {
	action := i.ObjectMeta.Labels["action"]

	if i.Status.State == "" {
		return false
	}

	return action == ActionUninstall && i.canUninstall()
}

func (i *Installation) canUninstall() bool {
	return (i.Status.State == StateInstalled || i.Status.State == StateError)
}

func (i *Installation) hasCondition(condition InstallationConditionType) bool {
	if i.Status.Conditions == nil || len(i.Status.Conditions) == 0 {
		return false
	}

	for _, f := range i.Status.Conditions {
		if f.Type == condition && f.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

func (i *Installation) CanBeDeleted() bool {
	result := i.Status.Conditions == nil || len(i.Status.Conditions) == 0 ||
		i.hasCondition(ConditionUninstalled) ||
		i.hasCondition(ConditionError)
	return result
}

func (i *Installation) IsBeingDeleted() bool {
	deletionTimestamp := i.GetDeletionTimestamp()
	return deletionTimestamp != nil
}

// InstallationSpec .
type InstallationSpec struct {
	KymaVersion string          `json:"version"`
	URL         string          `json:"url"`
	Components  []KymaComponent `json:"components"`
}

// KymaComponent represents single kyma component to be handled by the installer
type KymaComponent struct {
	Name        string `json:"name"`
	ReleaseName string `json:"release"`
	Namespace   string `json:"namespace"`
}

// GetReleaseName returns release name for component
func (kc KymaComponent) GetReleaseName() string {
	if len(kc.ReleaseName) > 0 {
		return kc.ReleaseName
	}

	return kc.Name
}

// StateEnum describes installation state
type StateEnum string

// InstallationConditionType defines installation condition type
type InstallationConditionType string

const (
	// StateEmpty .
	StateEmpty StateEnum = ""

	// StateInstalled means installation of kyma is done
	StateInstalled StateEnum = "Installed"

	// StateUninstalled means installation is removed without errors
	StateUninstalled StateEnum = "Uninstalled"

	// StateInProgress means installation/update/uninstallation is running
	StateInProgress StateEnum = "InProgress"

	// StateError means an error condition occurred during install/update/uninstall operation
	StateError StateEnum = "Error"

	// CondtitionInstalled .
	CondtitionInstalled InstallationConditionType = "Installed"

	// ConditionInstalling .
	ConditionInstalling InstallationConditionType = "Installing"

	// ConditionUninstalled .
	ConditionUninstalled InstallationConditionType = "Uninstalled"

	// ConditionUninstalling .
	ConditionUninstalling InstallationConditionType = "Uninstalling"

	// ConditionInProgress .
	ConditionInProgress InstallationConditionType = "InProgress"

	// ConditionError .
	ConditionError InstallationConditionType = "Error"

	// ActionInstall .
	ActionInstall string = "install"

	// ActionUninstall .
	ActionUninstall = "uninstall"
)

// InstallationCondition .
type InstallationCondition struct {
	Type               InstallationConditionType `json:"type"`
	Status             corev1.ConditionStatus    `json:"status"`
	LastTransitionTime metav1.Time               `json:"lastTransitionTime,omitempty"`
	LastProbeTime      metav1.Time               `json:"lastProbeTime,omitempty"`
}

// InstallationStatus .
type InstallationStatus struct {
	Conditions  []InstallationCondition `json:"conditions"`
	State       StateEnum               `json:"state"`
	Description string                  `json:"description"`
	KymaVersion string                  `json:"version"`
	URL         string                  `json:"url"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstallationList .
type InstallationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Installation `json:"items"`
}
