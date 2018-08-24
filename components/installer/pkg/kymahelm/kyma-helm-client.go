package kymahelm

import (
	"fmt"
	"log"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

// ClientInterface .
type ClientInterface interface {
	ListReleases() (*rls.ListReleasesResponse, error)
	ReleaseStatus(rname string) (string, error)
	InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error)
	InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error)
	InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error)
	UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error)
	DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error)
	PrintRelease(release *release.Release)
}

// Client .
type Client struct {
	helm *helm.Client
}

// NewClient .
func NewClient(host string) *Client {
	return &Client{
		helm: helm.NewClient(helm.Host(host)),
	}
}

// ListReleases .
func (hc *Client) ListReleases() (*rls.ListReleasesResponse, error) {
	return hc.helm.ListReleases()
}

//ReleaseStatus returns roughly-formatted Release status (columns are separated with blanks but not adjusted)
func (hc *Client) ReleaseStatus(rname string) (string, error) {
	status, err := hc.helm.ReleaseStatus(rname)
	if err != nil {
		return "", err
	}
	statusStr := fmt.Sprintf("%+v\n", status)
	return strings.Replace(statusStr, `\n`, "\n", -1), nil
}

// InstallReleaseFromChart .
func (hc *Client) InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error) {
	chart, err := chartutil.Load(chartdir)

	if err != nil {
		return nil, err
	}

	return hc.helm.InstallReleaseFromChart(
		chart,
		ns,
		helm.ReleaseName(string(releaseName)),
		helm.ValueOverrides([]byte(overrides)), //Without it default "values.yaml" file is ignored!
		helm.InstallWait(true),
		helm.InstallTimeout(3600),
	)
}

// InstallRelease .
func (hc *Client) InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	return hc.helm.InstallRelease(
		chartdir,
		ns,
		helm.ReleaseName(releasename),
		helm.ValueOverrides([]byte(overrides)), //Without it default "values.yaml" file is ignored!
		helm.InstallWait(true),
		helm.InstallTimeout(3600),
	)
}

// InstallReleaseWithoutWait .
func (hc *Client) InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	return hc.helm.InstallRelease(
		chartdir,
		ns,
		helm.ReleaseName(releasename),
		helm.ValueOverrides([]byte(overrides)), //Without it default "values.yaml" file is ignored!
		helm.InstallWait(false),
		helm.InstallTimeout(3600),
	)
}

// UpgradeRelease .
func (hc *Client) UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error) {
	return hc.helm.UpdateRelease(
		releaseName,
		chartDir,
		helm.UpdateValueOverrides([]byte(overrides+"\nglobal:\n  install: true\n")), //Without it default "values.yaml" file is ignored!
		helm.ReuseValues(true),
		helm.UpgradeTimeout(3600),
	)
}

// DeleteRelease .
func (hc *Client) DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error) {
	return hc.helm.DeleteRelease(
		releaseName,
		helm.DeletePurge(true),
		helm.DeleteTimeout(3600),
	)
}

//PrintRelease .
func (hc *Client) PrintRelease(release *release.Release) {
	log.Printf("Name: %s", release.Name)
	log.Printf("Namespace: %s", release.Namespace)
	log.Printf("Version: %d", release.Version)
	log.Printf("Status: %s", release.Info.Status.Code)
	log.Printf("Description: %s", release.Info.Description)
}
