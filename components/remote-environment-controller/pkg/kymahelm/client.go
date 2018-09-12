package kymahelm

import (
	"k8s.io/helm/pkg/helm"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type HelmClient interface {
	ListReleases() (*rls.ListReleasesResponse, error)
	InstallReleaseFromChart(chartDir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error)
	DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error)
}

type helmClient struct {
	helm *helm.Client
}

func NewClient(host string) HelmClient {
	return &helmClient{
		helm: helm.NewClient(helm.Host(host)),
	}
}

func (hc *helmClient) ListReleases() (*rls.ListReleasesResponse, error) {
	return hc.helm.ListReleases()
}

func (hc *helmClient) InstallReleaseFromChart(chartDir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error) {
	return hc.helm.InstallRelease(
		chartDir,
		ns,
		helm.ReleaseName(string(releaseName)),
		helm.ValueOverrides([]byte(overrides)), //Without it default "values.yaml" file is ignored!
		helm.InstallWait(true),
		helm.InstallTimeout(3600),
	)
}

func (hc *helmClient) DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error) {
	return hc.helm.DeleteRelease(
		releaseName,
		helm.DeletePurge(true),
		helm.DeleteTimeout(3600),
	)
}
