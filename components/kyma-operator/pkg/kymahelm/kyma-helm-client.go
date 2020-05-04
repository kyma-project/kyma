package kymahelm

import (
	"fmt"
	"log"
	"strings"

	"github.com/sirupsen/logrus"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/tlsutil"
)

// ClientInterface .
type ClientInterface interface {
	ListReleases() (*rls.ListReleasesResponse, error)
	ReleaseStatus(rname string) (string, error)
	IsReleaseDeletable(rname string) (bool, error)
	InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error)
	InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error)
	InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error)
	UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error)
	DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error)
	PrintRelease(release *release.Release)
}

// Client .
type Client struct {
	helm            *helm.Client
	overridesLogger *logrus.Logger
}

// NewClient .
func NewClient(host string, TLSKey string, TLSCert string, TLSInsecureSkipVerify bool, overridesLogger *logrus.Logger) (*Client, error) {
	tlsopts := tlsutil.Options{
		KeyFile:            TLSKey,
		CertFile:           TLSCert,
		InsecureSkipVerify: TLSInsecureSkipVerify,
	}
	tlscfg, err := tlsutil.ClientConfig(tlsopts)
	return &Client{
		helm:            helm.NewClient(helm.Host(host), helm.WithTLS(tlscfg), helm.ConnectTimeout(30)),
		overridesLogger: overridesLogger,
	}, err
}

// ListReleases lists all releases except for the superseded ones
func (hc *Client) ListReleases() (*rls.ListReleasesResponse, error) {
	statuses := []release.Status_Code{
		release.Status_DELETED,
		release.Status_DELETING,
		release.Status_DEPLOYED,
		release.Status_FAILED,
		release.Status_PENDING_INSTALL,
		release.Status_PENDING_ROLLBACK,
		release.Status_PENDING_UPGRADE,
		release.Status_UNKNOWN,
	}
	return hc.helm.ListReleases(helm.ReleaseListStatuses(statuses))
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

//IsReleaseDeletable returns true for release that can be deleted
func (hc *Client) IsReleaseDeletable(rname string) (bool, error){
	statusRes, err := hc.helm.ReleaseStatus(rname)
	if err != nil {
		return false, err
	}

	return statusRes.Info.Status.Code == release.Status_DEPLOYED || statusRes.Info.Status.Code == release.Status_FAILED, nil
}

// InstallReleaseFromChart .
func (hc *Client) InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error) {
	chart, err := chartutil.Load(chartdir)

	if err != nil {
		return nil, err
	}

	hc.PrintOverrides(overrides, releaseName, "installation")

	return hc.helm.InstallReleaseFromChart(
		chart,
		ns,
		helm.ReleaseName(string(releaseName)),
		helm.ValueOverrides([]byte(overrides)),
		helm.InstallWait(true),
		helm.InstallTimeout(3600),
	)
}

// InstallRelease .
func (hc *Client) InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	hc.PrintOverrides(overrides, releasename, "installation")

	return hc.helm.InstallRelease(
		chartdir,
		ns,
		helm.ReleaseName(releasename),
		helm.ValueOverrides([]byte(overrides)),
		helm.InstallWait(true),
		helm.InstallTimeout(3600),
	)
}

// InstallReleaseWithoutWait .
func (hc *Client) InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	hc.PrintOverrides(overrides, releasename, "installation")

	return hc.helm.InstallRelease(
		chartdir,
		ns,
		helm.ReleaseName(releasename),
		helm.ValueOverrides([]byte(overrides)),
		helm.InstallWait(false),
		helm.InstallTimeout(3600),
	)
}

// UpgradeRelease .
func (hc *Client) UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error) {
	hc.PrintOverrides(overrides, releaseName, "update")

	return hc.helm.UpdateRelease(
		releaseName,
		chartDir,
		helm.UpdateValueOverrides([]byte(overrides)),
		helm.ReuseValues(false),
		helm.UpgradeTimeout(3600),
		helm.UpgradeWait(true),
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

// PrintOverrides .
func (hc *Client) PrintOverrides(overrides string, releaseName string, action string) {
	hc.overridesLogger.Printf("Overrides used for %s of component %s", action, releaseName)

	if overrides == "" {
		hc.overridesLogger.Println("No overrides found")
		return
	}
	hc.overridesLogger.Println("\n", overrides)
}
