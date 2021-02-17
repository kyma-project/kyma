package kymahelm

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ClientInterface exposes functions to interact with Helm
type ClientInterface interface {
	ListReleases() ([]*Release, error)
	IsReleaseDeletable(nn NamespacedName) (bool, error)
	IsReleasePresent(nn NamespacedName) (bool, error)
	ReleaseDeployedRevision(nn NamespacedName) (int, error)
	InstallRelease(chartDir string, nn NamespacedName, overrides overrides.Map) (*Release, error)
	UpgradeRelease(chartDir string, nn NamespacedName, overrides overrides.Map) (*Release, error)
	UninstallRelease(nn NamespacedName) error
	RollbackRelease(nn NamespacedName, revision int) error
	WaitForReleaseDelete(nn NamespacedName) (bool, error)
	WaitForReleaseRollback(nn NamespacedName) (bool, error)
	PrintRelease(release *Release)
}

// Client is a Helm client
type Client struct {
	overridesLogger *logrus.Logger
	maxHistory      int
	driver          string
	logFn           func(string, ...interface{})
}

// NewClient initializes an instance of Client and returns it
func NewClient(overridesLogger *logrus.Logger, maxHistory int, driver string, debug bool) *Client {

	logFn := func(format string, v ...interface{}) {
		if debug {
			format = fmt.Sprintf("[debug] %s\n", format)
			log.Output(2, fmt.Sprintf(format, v...))
		}
	}

	return &Client{
		overridesLogger: overridesLogger,
		maxHistory:      maxHistory,
		driver:          driver,
		logFn:           logFn,
	}
}

// ListReleases lists all releases except for the superseded ones
func (hc *Client) ListReleases() ([]*Release, error) {

	cfg, err := hc.newActionConfig("")
	if err != nil {
		return nil, err
	}

	lister := action.NewList(cfg)
	lister.All = true
	lister.AllNamespaces = true

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

// ReleaseStatus returns release status
func (hc *Client) ReleaseStatus(nn NamespacedName) (*ReleaseStatus, error) {

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return nil, err
	}

	status := action.NewStatus(cfg)

	rel, err := status.Run(nn.Name)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(rel).ReleaseStatus, nil
}

// IsReleasePresent returns true for release that cis present on the cluster
func (hc *Client) IsReleasePresent(nn NamespacedName) (bool, error) {

	isPresent := true
	maxAttempts := 3
	fixedDelay := 3

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		if strings.Contains(err.Error(), driver.ErrReleaseNotFound.Error()) {
			return false, nil
		}
		return true, err
	}

	status := action.NewStatus(cfg)

	err = retry.Do(
		func() error {
			_, err := status.Run(nn.Name)
			if err != nil {
				if strings.Contains(err.Error(), driver.ErrReleaseNotFound.Error()) {
					isPresent = false
					return nil
				}
				return err
			}
			return nil
		},
		retry.Attempts(uint(maxAttempts)),
		retry.DelayType(func(attempt uint, config *retry.Config) time.Duration {
			log.Printf("Retry number %d on getting release status.\n", attempt+1)
			return time.Duration(fixedDelay) * time.Second
		}),
	)

	return isPresent, err
}

// IsReleaseDeletable returns true for release that can be deleted
func (hc *Client) IsReleaseDeletable(nn NamespacedName) (bool, error) {

	isDeletable := false
	maxAttempts := 3
	fixedDelay := 3

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return false, err
	}

	status := action.NewStatus(cfg)

	err = retry.Do(
		func() error {
			rel, err := status.Run(nn.Name)
			if err != nil {
				if strings.Contains(err.Error(), driver.ErrReleaseNotFound.Error()) {
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

// ReleaseDeployedRevision returns the last deployed revision
func (hc *Client) ReleaseDeployedRevision(nn NamespacedName) (int, error) {

	var deployedRevision = 0

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return deployedRevision, err
	}

	history := action.NewHistory(cfg)
	history.Max = hc.maxHistory

	relHistory, err := history.Run(nn.Name)
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

// InstallRelease installs a Helm chart
func (hc *Client) InstallRelease(chartDir string, nn NamespacedName, values overrides.Map) (*Release, error) {

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	install := action.NewInstall(cfg)
	install.ReleaseName = nn.Name
	install.Namespace = nn.Namespace
	install.Atomic = false
	install.Wait = true
	install.CreateNamespace = true

	hc.PrintOverrides(values, nn.Name, "install")

	installedRelease, err := install.Run(chart, values)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(installedRelease), nil
}

// UpgradeRelease upgrades a Helm chart
func (hc *Client) UpgradeRelease(chartDir string, nn NamespacedName, values overrides.Map) (*Release, error) {

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	upgrade := action.NewUpgrade(cfg)
	upgrade.Atomic = false
	upgrade.CleanupOnFail = true
	upgrade.Wait = true
	upgrade.ReuseValues = false
	upgrade.Recreate = false
	upgrade.MaxHistory = hc.maxHistory

	hc.PrintOverrides(values, nn.Name, "update")

	upgradedRelease, err := upgrade.Run(nn.Name, chart, values)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(upgradedRelease), nil
}

//RollbackRelease performs rollback of a Helm release
func (hc *Client) RollbackRelease(nn NamespacedName, revision int) error {

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return err
	}

	rollback := action.NewRollback(cfg)
	rollback.Wait = true
	rollback.Version = revision
	rollback.CleanupOnFail = true
	rollback.Recreate = true

	return rollback.Run(nn.Name)
}

// UninstallRelease uninstalls a Helm release
func (hc *Client) UninstallRelease(nn NamespacedName) error {

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return err
	}

	uninstall := action.NewUninstall(cfg)

	_, err = uninstall.Run(nn.Name)

	return err
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

func (hc *Client) newActionConfig(namespace string) (*action.Configuration, error) {
	clientGetter := genericclioptions.NewConfigFlags(false)
	clientGetter.Namespace = &namespace

	cfg := new(action.Configuration)
	if err := cfg.Init(clientGetter, namespace, hc.driver, hc.logFn); err != nil {
		return nil, err
	}

	return cfg, nil
}
