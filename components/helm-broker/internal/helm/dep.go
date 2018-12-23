package helm

import (
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type helmDeleteInstaller interface {
	ListReleases(opts ...helm.ReleaseListOption) (*rls.ListReleasesResponse, error)
	InstallReleaseFromChart(chart *chart.Chart, ns string, opts ...helm.InstallOption) (*rls.InstallReleaseResponse, error)
	UpdateReleaseFromChart(rlsName string, chart *chart.Chart, opts ...helm.UpdateOption) (*rls.UpdateReleaseResponse, error)
	DeleteRelease(rlsName string, opts ...helm.DeleteOption) (*rls.UninstallReleaseResponse, error)
}
