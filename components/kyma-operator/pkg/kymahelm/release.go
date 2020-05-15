package kymahelm

import (
	"errors"
	"fmt"
	helm "k8s.io/helm/pkg/proto/hapi/release"
)

// ReleaseStatus is an inner representation of a Helm release
type ReleaseStatus struct {
	StatusCode           helm.Status_Code
	CurrentRevision      int32
	LastDeployedRevision int32
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
