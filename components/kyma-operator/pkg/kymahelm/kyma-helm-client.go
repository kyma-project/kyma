package kymahelm

import (
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"log"

	"k8s.io/helm/pkg/helm"
)

// ClientInterface .
type ClientInterface interface {
	ListReleases() ([]*Release, error)
	ReleaseStatus(name string) (string, error)
	IsReleaseDeletable(rname string) (bool, error)
	ReleaseDeployedRevision(rname string) (int32, error)
	InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*Release, error)
	InstallRelease(chartdir, ns, releasename, overrides string) (*Release, error)
	InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*Release, error)
	UpgradeRelease(chartDir, releaseName, overrides string) (*Release, error)
	DeleteRelease(releaseName string) (*Release, error) //todo: rename to "uninstall"
	RollbackRelease(releaseName string, revision int32) (*Release, error)
	PrintRelease(release *Release)
}

// Client .
type Client struct {
	cfg             *action.Configuration
	helm            *helm.Client
	overridesLogger *logrus.Logger
	maxHistory      int32
	timeout         int64
}

// NewClient .
func NewClient(host string, TLSKey string, TLSCert string, TLSInsecureSkipVerify bool, overridesLogger *logrus.Logger, maxHistory int32, timeout int64) (*Client, error) { //keeping the signature; todo: remove unused params,  change int32 -> int, don't return error
	return &Client{
		cfg:             nil, //todo: pass in params
		overridesLogger: overridesLogger,
		maxHistory:      maxHistory,
		timeout:         timeout,
	}, nil
}

// ListReleases lists all releases except for the superseded ones
func (hc *Client) ListReleases() ([]*Release, error) {

	lister := action.NewList(hc.cfg)
	lister.All = true
	//todo: sorter?

	releases, err := lister.Run()
	if err != nil {
		return nil, err
	}

	var kymaReleases []*Release

	for _, hr := range releases {
		kymaReleases = append(kymaReleases, helmReleaseToKymaRelease(hr))
	}

	return kymaReleases, nil
}

//ReleaseStatus returns roughly-formatted Release status (columns are separated with blanks but not adjusted)
func (hc *Client) ReleaseStatus(name string) (string, error) {

	status := action.NewStatus(hc.cfg)
	//status.Version = 0 // default: 0 -> get last

	rel, err := status.Run(name)
	if err != nil {
		return "", err
	}

	return rel.Info.Status.String(), nil
}

//IsReleaseDeletable returns true for release that can be deleted
func (hc *Client) IsReleaseDeletable(rname string) (bool, error) {
	//isDeletable := false
	//maxAttempts := 3
	//fixedDelay := 3
	//
	//err := retry.Do(
	//	func() error {
	//		statusRes, err := hc.helm.ReleaseStatus(rname)
	//		if err != nil {
	//			if strings.Contains(err.Error(), errors.ErrReleaseNotFound(rname).Error()) {
	//				isDeletable = false
	//				return nil
	//			}
	//			return err
	//		}
	//		isDeletable = statusRes.Info.Status.Code != release.Status_DEPLOYED
	//		return nil
	//	},
	//	retry.Attempts(uint(maxAttempts)),
	//	retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
	//		log.Printf("Retry number %d on getting release status.\n", attempt+1)
	//		return time.Duration(fixedDelay) * time.Second
	//	}),
	//)
	//
	//return isDeletable, err

	return true, nil
}

func (hc *Client) ReleaseDeployedRevision(rname string) (int32, error) {

	//var deployedRevision int32 = 0
	//
	//releaseHistoryRes, err := hc.helm.ReleaseHistory(rname, helm.WithMaxHistory(int32(hc.maxHistory)))
	//if err != nil {
	//	return deployedRevision, err
	//}
	//
	//for _, rel := range releaseHistoryRes.Releases {
	//	if rel.Info.Status.Code == release.Status_DEPLOYED {
	//		deployedRevision = rel.Version
	//	}
	//}

	return 0, nil
}

// InstallReleaseFromChart .
func (hc *Client) InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*Release, error) {

	// Helm3 install release

	return nil, nil
}

// InstallRelease .
func (hc *Client) InstallRelease(chartdir, ns, releasename, overrides string) (*Release, error) {

	// Helm3 install release

	return nil, nil
}

// InstallReleaseWithoutWait .
func (hc *Client) InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*Release, error) {

	// Helm3 install release

	return nil, nil
}

// UpgradeRelease .
func (hc *Client) UpgradeRelease(chartDir, releaseName, overrides string) (*Release, error) {

	// Helm3 upgrade release

	return nil, nil
}

//RollbackRelease performs rollback to given revision
func (hc *Client) RollbackRelease(releaseName string, revision int32) (*Release, error) {

	// Helm3 rollback release

	return nil, nil
}

// DeleteRelease .
func (hc *Client) DeleteRelease(releaseName string) (*Release, error) { //todo: rename to "uninstall"

	// Helm3 uninstall release

	return nil, nil
}

//PrintRelease .
func (hc *Client) PrintRelease(release *Release) {
	log.Printf("Name: %s", release.Name)
	log.Printf("Namespace: %s", release.Namespace)
	log.Printf("Version: %d", release.CurrentRevision)
	log.Printf("Status: %s", release.StatusCode)
	log.Printf("Description: %s", release.Description)
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
