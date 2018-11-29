package testkit

import (
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"time"
)

type HelmClient interface {
	CheckReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error)
	CheckReleaseExistence(name string) (bool, error)
}

type helmClient struct {
	helm          *helm.Client
	retryCount    int
	retryWaitTime time.Duration
}

func NewHelmClient(host string) HelmClient {
	return &helmClient{
		helm: helm.NewClient(helm.Host(host)),
	}
}

func (hc *helmClient) CheckReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error) {
	return hc.helm.ReleaseStatus(rlsName)
}

func (hc *helmClient) CheckReleaseExistence(name string) (bool, error) {
	listResponse, err := hc.helm.ListReleases(helm.ReleaseListStatuses([]release.Status_Code{
		release.Status_DELETED,
		release.Status_DELETING,
		release.Status_DEPLOYED,
		release.Status_FAILED,
		release.Status_PENDING_INSTALL,
		release.Status_PENDING_ROLLBACK,
		release.Status_PENDING_UPGRADE,
		release.Status_SUPERSEDED,
		release.Status_UNKNOWN,
	}))
	if err != nil {
		return false, err
	}

	for _, rel := range listResponse.Releases {
		if rel.Name == name {
			return true, nil
		}
	}
	return false, nil
}
