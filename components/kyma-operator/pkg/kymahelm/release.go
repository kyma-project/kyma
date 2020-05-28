package kymahelm

import (
	"errors"
	"fmt"

	helm "k8s.io/helm/pkg/proto/hapi/release"
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
	StatusCode           helm.Status_Code
	CurrentRevision      int32
	LastDeployedRevision int32
}

// UninstallReleaseResponse is an internal representation of Helm's uninstall release response
type UninstallReleaseStatus struct {
}

func (rs *ReleaseStatus) IsUpgradeStep() (bool, error) {

	switch rs.StatusCode {

	case helm.Status_PENDING_INSTALL:
		return false, nil

	case helm.Status_DEPLOYED, helm.Status_PENDING_UPGRADE, helm.Status_PENDING_ROLLBACK:
		return true, nil

	case helm.Status_FAILED, helm.Status_UNKNOWN, helm.Status_DELETED, helm.Status_DELETING:

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
