package kymahelm

import (
	"errors"
	"fmt"

	"helm.sh/helm/v3/pkg/release"
)

type Status string

const (
	// StatusUnknown indicates that a release is in an uncertain state.
	StatusUnknown Status = "unknown"
	// StatusDeployed indicates that the release has been pushed to Kubernetes.
	StatusDeployed Status = "deployed"
	// StatusUninstalled indicates that a release has been uninstalled from Kubernetes.
	StatusUninstalled Status = "uninstalled"
	// StatusSuperseded indicates that this release object is outdated and a newer one exists.
	StatusSuperseded Status = "superseded"
	// StatusFailed indicates that the release was not successfully deployed.
	StatusFailed Status = "failed"
	// StatusUninstalling indicates that a uninstall operation is underway.
	StatusUninstalling Status = "uninstalling"
	// StatusPendingInstall indicates that an install operation is underway.
	StatusPendingInstall Status = "pending-install"
	// StatusPendingUpgrade indicates that an upgrade operation is underway.
	StatusPendingUpgrade Status = "pending-upgrade"
	// StatusPendingRollback indicates that an rollback operation is underway.
	StatusPendingRollback Status = "pending-rollback"
)

// Release is an internal representation of a Helm release
type Release struct {
	*ReleaseMeta
	*ReleaseStatus
}

// NamespacedName Combines release name and namespace
type NamespacedName struct {
	Name      string
	Namespace string
}

// ReleaseMeta is an internal representation of Helm's release metadata
type ReleaseMeta struct {
	NamespacedName
	Description string
}

// ReleaseStatus is an internal representation of Helm's release status
type ReleaseStatus struct {
	Status               Status
	CurrentRevision      int
	LastDeployedRevision int
}

// UninstallReleaseResponse is an internal representation of Helm's uninstall release response
type UninstallReleaseStatus struct {
}

func helmReleaseToKymaRelease(hr *release.Release) *Release {
	return &Release{
		&ReleaseMeta{
			NamespacedName: NamespacedName{Name: hr.Name, Namespace: hr.Namespace},
			Description:    hr.Info.Description,
		},
		&ReleaseStatus{
			Status:          Status(hr.Info.Status),
			CurrentRevision: hr.Version,
		},
	}
}

func (rs *ReleaseStatus) IsUpgradeStep() (bool, error) {

	switch rs.Status {

	case StatusPendingInstall:
		return false, nil

	case StatusDeployed, StatusPendingUpgrade, StatusPendingRollback:
		return true, nil

	case StatusFailed, StatusUnknown, StatusUninstalled, StatusUninstalling:

		if rs.hasMultipleRevisions() {

			if rs.isDeployed() {
				return true, nil
			}

			return false, errors.New("no deployed revision to rollback to")
		}

		return false, nil

	default:
		return false, errors.New(fmt.Sprintf("unexpected status %s", rs.Status))
	}
}

func (rs *ReleaseStatus) isDeployed() bool {
	return rs.LastDeployedRevision > 0
}

func (rs *ReleaseStatus) hasMultipleRevisions() bool {
	return rs.CurrentRevision > 1
}
