package kymahelm

import (
	"log"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ClientInterface .
type ClientInterface interface {
	ListReleases() ([]*Release, error)
	ReleaseStatus(relName string) (string, error)
	IsReleaseDeletable(relName string) (bool, error)
	ReleaseDeployedRevision(relName string) (int, error)
	InstallReleaseFromChart(chartDir, ns, relName string, overrides overrides.Map) (*Release, error)
	InstallRelease(chartDir, ns, relName string, overrides overrides.Map) (*Release, error)
	InstallReleaseWithoutWait(chartDir, ns, relName string, overrides overrides.Map) (*Release, error)
	UpgradeRelease(chartDir, relName string, overrides overrides.Map) (*Release, error)
	DeleteRelease(relName string) (*Release, error) //todo: rename to "uninstall"
	RollbackRelease(relName string, revision int) (*Release, error)
	PrintRelease(release *Release)
}

// Client .
type Client struct {
	cfg             *action.Configuration
	overridesLogger *logrus.Logger
	maxHistory      int
	timeout         time.Duration
}

// NewClient .
func NewClient(overridesLogger *logrus.Logger, maxHistory int, timeout int64) (*Client, error) {
	cfg := &action.Configuration{}
	clientGetter := genericclioptions.NewConfigFlags(false)
	debugLog := overridesLogger.Printf

	if err := cfg.Init(clientGetter, "kyma-installer", "memory", debugLog); err != nil {
		return nil, err
	}

	return &Client{
		cfg:             cfg,
		overridesLogger: overridesLogger,
		maxHistory:      maxHistory,
		timeout:         time.Duration(timeout) * time.Second,
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
func (hc *Client) ReleaseStatus(relName string) (string, error) {

	status := action.NewStatus(hc.cfg)
	//status.Version = 0 // default: 0 -> get last

	rel, err := status.Run(relName)
	if err != nil {
		return "", err
	}

	return rel.Info.Status.String(), nil
}

//IsReleaseDeletable returns true for release that can be deleted
func (hc *Client) IsReleaseDeletable(relName string) (bool, error) { //todo: helm3 allows atomic operations, this func might be useless

	isDeletable := false
	maxAttempts := 3
	fixedDelay := 3

	status := action.NewStatus(hc.cfg)

	err := retry.Do(
		func() error {
			rel, err := status.Run(relName)
			if err != nil {
				if strings.Contains(err.Error(), "not found") { //todo: replace with actual h3 error if it exists
					isDeletable = false
					return nil
				}
				return err
			}
			isDeletable = rel.Info.Status != release.StatusDeployed
			return nil
		},
		retry.Attempts(uint(maxAttempts)),
		retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
			log.Printf("Retry number %d on getting release status.\n", attempt+1)
			return time.Duration(fixedDelay) * time.Second
		}),
	)

	return isDeletable, err
}

func (hc *Client) ReleaseDeployedRevision(relName string) (int, error) { //todo: helm3 allows atomic operations, this func might be useless

	var deployedRevision = 0

	history := action.NewHistory(hc.cfg)
	history.Max = hc.maxHistory

	relHistory, err := history.Run(relName)
	if err != nil {
		return deployedRevision, err
	}

	for _, rel := range relHistory {
		if rel.Info.Status == release.StatusDeployed {
			deployedRevision = rel.Version
			break
		}
	}

	return deployedRevision, nil
}

// InstallReleaseFromChart .
func (hc *Client) InstallReleaseFromChart(chartDir, ns, relName string, values overrides.Map) (*Release, error) {

	chart, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	install := action.NewInstall(hc.cfg)
	install.ReleaseName = relName
	install.Namespace = ns
	install.Atomic = true
	install.Wait = true
	install.Timeout = hc.timeout
	install.CreateNamespace = true //https://v3.helm.sh/docs/faq/#automatically-creating-namespaces

	hc.PrintOverrides(values, relName, "install")

	installedRelease, err := install.Run(chart, values)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(installedRelease), nil
}

// InstallRelease .
func (hc *Client) InstallRelease(chartDir, ns, relName string, values overrides.Map) (*Release, error) {
	return hc.InstallReleaseFromChart(chartDir, ns, relName, values)
}

// InstallReleaseWithoutWait .
func (hc *Client) InstallReleaseWithoutWait(chartDir, ns, relName string, values overrides.Map) (*Release, error) { //todo: implemented with wait, we don't need that function anyways
	return hc.InstallReleaseFromChart(chartDir, ns, relName, values)
}

// UpgradeRelease .
func (hc *Client) UpgradeRelease(chartDir, relName string, values overrides.Map) (*Release, error) {

	chart, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	upgrade := action.NewUpgrade(hc.cfg)
	upgrade.Atomic = true
	upgrade.CleanupOnFail = true
	upgrade.Wait = true
	upgrade.Timeout = hc.timeout
	upgrade.ReuseValues = true

	hc.PrintOverrides(values, relName, "update")

	upgradedRelease, err := upgrade.Run(relName, chart, values)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(upgradedRelease), nil
}

//RollbackRelease performs rollback to given revision
func (hc *Client) RollbackRelease(relName string, revision int) (*Release, error) {

	rollback := action.NewRollback(hc.cfg)
	rollback.Wait = true
	rollback.Timeout = hc.timeout
	rollback.Version = revision
	rollback.CleanupOnFail = true
	rollback.Recreate = true

	return nil, rollback.Run(relName) //todo: return only error or fetch actual object
}

// DeleteRelease .
func (hc *Client) DeleteRelease(relName string) (*Release, error) { //todo: rename to "uninstall"

	uninstall := action.NewUninstall(hc.cfg)
	uninstall.Timeout = hc.timeout

	_, err := uninstall.Run(relName)
	if err != nil {
		return nil, err
	}

	return &Release{}, nil //todo: return only error or transform uninstall response to internal type or I don't care rly
}

//PrintRelease .
func (hc *Client) PrintRelease(release *Release) {
	log.Printf("Name: %s", release.Name)
	log.Printf("Namespace: %s", release.Namespace)
	log.Printf("Version: %d", release.CurrentRevision)
	log.Printf("Status: %s", release.Status)
	log.Printf("Description: %s", release.Description)
}

// PrintOverrides .
func (hc *Client) PrintOverrides(values overrides.Map, relName string, action string) {

	hc.overridesLogger.Printf("Overrides used to %s component %s", action, relName)

	if len(values) == 0 {
		hc.overridesLogger.Println("No overrides found")
		return
	}

	hc.overridesLogger.Println(overrides.ToYaml(values))
}
