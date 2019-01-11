package kymahelm

import (
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type HelmClient interface {
	ListReleases() (*rls.ListReleasesResponse, error)
	InstallReleaseFromChart(chartDir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error)
	DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error)
	ReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error)
}

type helmClient struct {
	helm                *helm.Client
	installationTimeout int64
}

func NewClient(host string, installationTimeout int64) HelmClient {
	return &helmClient{
		helm:                helm.NewClient(helm.Host(host)),
		installationTimeout: installationTimeout,
	}
}

func (hc *helmClient) ListReleases() (*rls.ListReleasesResponse, error) {
	statuses := []release.Status_Code{
		release.Status_DELETED,
		release.Status_DELETING,
		release.Status_DEPLOYED,
		release.Status_FAILED,
		release.Status_PENDING_INSTALL,
		release.Status_PENDING_ROLLBACK,
		release.Status_PENDING_UPGRADE,
		release.Status_SUPERSEDED,
		release.Status_UNKNOWN,
	}

	return hc.helm.ListReleases(helm.ReleaseListStatuses(statuses))
}

func (hc *helmClient) InstallReleaseFromChart(chartDir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error) {
	return hc.helm.InstallRelease(
		chartDir,
		ns,
		helm.ReleaseName(string(releaseName)),
		helm.ValueOverrides([]byte(overrides)), //Without it default "values.yaml" file is ignored!
		helm.InstallWait(true),
		helm.InstallTimeout(hc.installationTimeout),
	)
}

func (hc *helmClient) DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error) {
	return hc.helm.DeleteRelease(
		releaseName,
		helm.DeletePurge(true),
		helm.DeleteTimeout(hc.installationTimeout),
	)
}

func (hc *helmClient) ReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error) {
	return hc.helm.ReleaseStatus(rlsName)
}
