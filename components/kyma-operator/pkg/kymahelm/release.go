package kymahelm

import (
	"errors"
	"fmt"
)

type Status_Code int32

const (
	// Status_UNKNOWN indicates that a release is in an uncertain state.
	Status_UNKNOWN Status_Code = 0
	// Status_DEPLOYED indicates that the release has been pushed to Kubernetes.
	Status_DEPLOYED Status_Code = 1
	// Status_DELETED indicates that a release has been deleted from Kubernetes.
	Status_DELETED Status_Code = 2
	// Status_SUPERSEDED indicates that this release object is outdated and a newer one exists.
	Status_SUPERSEDED Status_Code = 3
	// Status_FAILED indicates that the release was not successfully deployed.
	Status_FAILED Status_Code = 4
	// Status_DELETING indicates that a delete operation is underway.
	Status_DELETING Status_Code = 5
	// Status_PENDING_INSTALL indicates that an install operation is underway.
	Status_PENDING_INSTALL Status_Code = 6
	// Status_PENDING_UPGRADE indicates that an upgrade operation is underway.
	Status_PENDING_UPGRADE Status_Code = 7
	// Status_PENDING_ROLLBACK indicates that a rollback operation is underway.
	Status_PENDING_ROLLBACK Status_Code = 8
)

//Release is an insternal representation of a Helm release
type Release struct {
	*ReleaseMeta
	*ReleaseStatus
}

//ReleaseMeta is an internal representation of Helm's release metadata
type ReleaseMeta struct {
	Name        string
	Namespace   string
	Description string
}

// ReleaseStatus is an internal representation of Helm's release status
type ReleaseStatus struct {
	StatusCode           Status_Code
	CurrentRevision      int32
	LastDeployedRevision int32
}

// UninstallReleaseResponse is an internal representation of Helm's uninstall release response
type UninstallReleaseStatus struct {
}

func (rs *ReleaseStatus) IsUpgradeStep() (bool, error) {

	switch rs.StatusCode {

	case Status_PENDING_INSTALL:
		return false, nil

	case Status_DEPLOYED, Status_PENDING_UPGRADE, Status_PENDING_ROLLBACK:
		return true, nil

	case Status_FAILED, Status_UNKNOWN, Status_DELETED, Status_DELETING:

		if rs.hasMultipleRevisions() {

			if rs.isDeployed() {
				return true, nil
			}

			return false, errors.New("no deployed revision to rollback to")
		}

		return false, nil

	default:
		return false, errors.New(fmt.Sprintf("unexpected status code %s", rs.StatusCode))
	}
}

func (rs *ReleaseStatus) isDeployed() bool {
	return rs.LastDeployedRevision > 0
}

func (rs *ReleaseStatus) hasMultipleRevisions() bool {
	return rs.CurrentRevision > 1
}
