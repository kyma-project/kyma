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

//todo: pass relname and relnamespace in a dedicated internal structure namespacedName

//todo: each function in a separate file?

// ClientInterface .
type ClientInterface interface {
	ListReleases() ([]*Release, error)
	IsReleaseDeletable(relNamespace, relName string) (bool, error)
	ReleaseDeployedRevision(relNamespace, relName string) (int, error)
	InstallReleaseFromChart(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	InstallRelease(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	InstallReleaseWithoutWait(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	UpgradeRelease(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	DeleteRelease(relNamespace, relName string) (*Release, error) //todo: rename to "uninstall"
	RollbackRelease(relNamespace, relName string, revision int) (*Release, error)
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

	cfg, err := hc.newActionConfig("") //todo: is that ok???????
	if err != nil {
		return nil, err
	}

	lister := action.NewList(cfg)
	lister.All = true
	lister.AllNamespaces = true //todo: is that ok?
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

// ReleaseStatus returns release status
func (hc *Client) ReleaseStatus(nn NamespacedName) (*ReleaseStatus, error) {

	cfg, err := hc.newActionConfig(nn.Namespace)
	if err != nil {
		return nil, err
	}

	status := action.NewStatus(cfg)
	//status.Version = 0 // default: 0 -> get last

	rel, err := status.Run(nn.Name)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(rel).ReleaseStatus, nil
}

//IsReleaseDeletable returns true for release that can be deleted
func (hc *Client) IsReleaseDeletable(relNamespace, relName string) (bool, error) { //todo: helm3 allows atomic operations, this func might be useless

	isDeletable := false
	maxAttempts := 3
	fixedDelay := 3

	cfg, err := hc.newActionConfig(relNamespace)
	if err != nil {
		return false, err
	}

	status := action.NewStatus(cfg)

	err = retry.Do(
		func() error {
			rel, err := status.Run(relName)
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

func (hc *Client) ReleaseDeployedRevision(relNamespace, relName string) (int, error) { //todo: helm3 allows atomic operations, this func might be useless

	var deployedRevision = 0

	cfg, err := hc.newActionConfig(relNamespace)
	if err != nil {
		return deployedRevision, err
	}

	history := action.NewHistory(cfg)
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
func (hc *Client) InstallReleaseFromChart(chartDir, relNamespace, relName string, values overrides.Map) (*Release, error) {

	cfg, err := hc.newActionConfig(relNamespace) //todo: parameterize driver
	//cfg, err := hc.newActionConfigMst(relNamespace)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	install := action.NewInstall(cfg) //todo: stretch: implement configurator, see https://github.com/fluxcd/helm-operator/blob/706bcb34841ed65fed007ad706082f28429e19bb/pkg/helm/v3/upgrade.go#L52
	install.ReleaseName = relName
	install.Namespace = relNamespace
	install.Atomic = false
	install.Wait = true            //todo: defaults to true if atomic is set. Remove if atomic == true
	install.CreateNamespace = true // see https://v3.helm.sh/docs/faq/#automatically-creating-namespaces

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
func (hc *Client) UpgradeRelease(chartDir, relNamespace, relName string, values overrides.Map) (*Release, error) {

	cfg, err := hc.newActionConfig(relNamespace)
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

	hc.PrintOverrides(values, relName, "update")

	upgradedRelease, err := upgrade.Run(relName, chart, values)
	if err != nil {
		return nil, err
	}

	return helmReleaseToKymaRelease(upgradedRelease), nil
}

//RollbackRelease performs rollback to given revision
func (hc *Client) RollbackRelease(relNamespace, relName string, revision int) (*Release, error) {

	cfg, err := hc.newActionConfig(relNamespace)
	if err != nil {
		return nil, err
	}

	rollback := action.NewRollback(cfg)
	rollback.Wait = true
	rollback.Version = revision
	rollback.CleanupOnFail = true
	rollback.Recreate = true

	return nil, rollback.Run(relName) //todo: return only error or fetch actual object
}

// DeleteRelease .
func (hc *Client) DeleteRelease(relNamespace, relName string) (*Release, error) { //todo: rename to "uninstall"

	cfg, err := hc.newActionConfig(relNamespace)
	if err != nil {
		return nil, err
	}

	uninstall := action.NewUninstall(cfg)

	_, err = uninstall.Run(relName)
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

func (hc *Client) newActionConfig(namespace string) (*action.Configuration, error) {
	clientGetter := genericclioptions.NewConfigFlags(false)
	clientGetter.Namespace = &namespace

	cfg := new(action.Configuration)
	if err := cfg.Init(clientGetter, namespace, hc.driver, hc.logFn); err != nil {
		return nil, err
	}

	return cfg, nil
}
