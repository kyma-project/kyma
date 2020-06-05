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

//todo: each function in a separate file?

// ClientInterface .
type ClientInterface interface {
	ListReleases() ([]*Release, error)
	IsReleaseDeletable(nn NamespacedName) (bool, error)
	ReleaseDeployedRevision(nn NamespacedName) (int, error)
	InstallReleaseFromChart(chartDir string, nn NamespacedName, overrides overrides.Map) (*Release, error)
	InstallRelease(chartDir string, nn NamespacedName, overrides overrides.Map) (*Release, error)
	UpgradeRelease(chartDir string, nn NamespacedName, overrides overrides.Map) (*Release, error)
	UninstallRelease(nn NamespacedName) error
	RollbackRelease(nn NamespacedName, revision int) error
	WaitForReleaseDelete(nn NamespacedName) (bool, error)
	WaitForReleaseRollback(nn NamespacedName) (bool, error)
	PrintRelease(release *Release)
}

type infoLogFunc func(string, ...interface{})

// Client .
type Client struct {
	overridesLogger *logrus.Logger
	maxHistory      int
	timeout         time.Duration //todo: timeout param consumed by actions limits single applies rather than entire operations (helm install, helm upgrade, etc.). Either remove or find a workaround
	driver          string        //todo: add validation in main?
	logFn           func(string, ...interface{})
}

// NewClient .
func NewClient(overridesLogger *logrus.Logger, maxHistory int, timeout int64, driver string, debug bool) *Client {

	logFn := func(format string, v ...interface{}) {
		if debug {
			format = fmt.Sprintf("[debug] %s\n", format)
			log.Output(2, fmt.Sprintf(format, v...))
		}
	}

	return &Client{
		overridesLogger: overridesLogger,
		maxHistory:      maxHistory,
		timeout:         time.Duration(timeout) * time.Second,
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

//IsReleaseDeletable returns true for release that can be deleted
func (hc *Client) IsReleaseDeletable(nn NamespacedName) (bool, error) { //todo: helm3 allows atomic operations, this func might be useless

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

func (hc *Client) ReleaseDeployedRevision(nn NamespacedName) (int, error) { //todo: helm3 allows atomic operations, this func might be useless

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

// InstallReleaseFromChart .
func (hc *Client) InstallReleaseFromChart(chartDir string, nn NamespacedName, values overrides.Map) (*Release, error) {

	cfg, err := hc.newActionConfig(nn.Namespace)
	//cfg, err := hc.newActionConfigMst(relNamespace)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	install := action.NewInstall(cfg) //todo: stretch: implement configurator
	install.ReleaseName = nn.Name
	install.Namespace = nn.Namespace
	install.Atomic = false
	install.Wait = true            //todo: defaults to true if atomic is set. Remove if atomic == true
	install.CreateNamespace = true // see https://v3.helm.sh/docs/faq/#automatically-creating-namespaces

	hc.PrintOverrides(values, nn.Name, "install")

	installedRelease, err := install.Run(chart, values)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(installedRelease), nil
}

// InstallRelease .
func (hc *Client) InstallRelease(chartDir string, nn NamespacedName, values overrides.Map) (*Release, error) {
	return hc.InstallReleaseFromChart(chartDir, nn, values)
}

// UpgradeRelease .
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
	upgrade.ReuseValues = true
	upgrade.Recreate = true

	hc.PrintOverrides(values, nn.Name, "update")

	upgradedRelease, err := upgrade.Run(nn.Name, chart, values)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(upgradedRelease), nil
}

//RollbackRelease performs rollback to given revision
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

// DeleteRelease uninstall a given release
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
