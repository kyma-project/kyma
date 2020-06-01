package kymahelm

import (
	"fmt"
	"log"
	"strings"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"

	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

//todo: pass relname and relnamespace in a dedicated internal structure namespacedName

//todo: each function in a separate file?

// ClientInterface .
type ClientInterface interface {
	ListReleases() ([]*Release, error)
	ReleaseStatus(relNamespace, relName string) (string, error)
	IsReleaseDeletable(relNamespace, relName string) (bool, error)
	ReleaseDeployedRevision(relNamespace, relName string) (int, error)
	InstallReleaseFromChart(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	InstallRelease(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	InstallReleaseWithoutWait(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	UpgradeRelease(chartDir, relNamespace, relName string, overrides overrides.Map) (*Release, error)
	DeleteRelease(relNamespace, relName string) (*Release, error) //todo: rename to "uninstall"
	RollbackRelease(relNamespace, relName string, revision int) (*Release, error)
	PrintRelease(release *Release)
}

type infoLogFunc func(string, ...interface{})

// Client .
type Client struct {
	kubeConfig      *rest.Config
	overridesLogger *logrus.Logger
	maxHistory      int
	timeout         time.Duration //todo: timeout param consumed by actions limits single applies rather than entire operations (helm install, helm upgrade, etc.). Either remove or find a workaround
}

func (hc *Client) infoLogFunc(namespace string, releaseName string) infoLogFunc {
	return func(format string, args ...interface{}) {
		message := fmt.Sprintf(format, args...)
		log.Printf("info: %s, targetNamespace: %s, release: %s", message, namespace, releaseName)
	}
}

// NewClient .
func NewClient(kubeConfig *rest.Config, overridesLogger *logrus.Logger, maxHistory int, timeout int64) (*Client, error) {

	return &Client{
		kubeConfig:      kubeConfig,
		overridesLogger: overridesLogger,
		maxHistory:      maxHistory,
		timeout:         time.Duration(timeout) * time.Second,
	}, nil
}

// ListReleases lists all releases except for the superseded ones
func (hc *Client) ListReleases() ([]*Release, error) {

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc("all", "all"), "", "") //todo: is that ok???????
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

//ReleaseStatus returns roughly-formatted Release status (columns are separated with blanks but not adjusted)
func (hc *Client) ReleaseStatus(relNamespace, relName string) (string, error) {

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc(relNamespace, relName), relNamespace, "")
	if err != nil {
		return "", err
	}

	status := action.NewStatus(cfg)
	//status.Version = 0 // default: 0 -> get last

	rel, err := status.Run(relName)
	if err != nil {
		return "", err
	}

	return rel.Info.Status.String(), nil
}

//IsReleaseDeletable returns true for release that can be deleted
func (hc *Client) IsReleaseDeletable(relNamespace, relName string) (bool, error) { //todo: helm3 allows atomic operations, this func might be useless

	isDeletable := false
	maxAttempts := 3
	fixedDelay := 3

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc(relNamespace, relName), relNamespace, "")
	if err != nil {
		return false, err
	}

	status := action.NewStatus(cfg)

	err = retry.Do(
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

func (hc *Client) ReleaseDeployedRevision(relNamespace, relName string) (int, error) { //todo: helm3 allows atomic operations, this func might be useless

	var deployedRevision = 0

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc(relNamespace, relName), relNamespace, "")
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

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc(relNamespace, relName), relNamespace, "") //todo: parameterize driver
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
	install.Wait = true //todo: defaults to true if atomic is set. Remove if atomic == true
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

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc(relNamespace, relName), relNamespace, "")
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	upgrade := action.NewUpgrade(cfg)
	upgrade.Atomic = true
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

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc(relNamespace, relName), relNamespace, "")
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

	cfg, err := newActionConfig(hc.kubeConfig, hc.infoLogFunc(relNamespace, relName), relNamespace, "")
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

func newActionConfig(config *rest.Config, logFunc infoLogFunc, namespace, driver string) (*action.Configuration, error) {

	restClientGetter := newConfigFlags(config, namespace)
	kubeClient := &kube.Client{
		Factory: util.NewFactory(restClientGetter),
		Log:     logFunc,
	}
	client, err := kubeClient.Factory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	store, err := newStorageDriver(client, logFunc, namespace, driver)
	if err != nil {
		return nil, err
	}

	return &action.Configuration{
		RESTClientGetter: restClientGetter,
		Releases:         store,
		KubeClient:       kubeClient,
		Log:              logFunc,
	}, nil
}

func newConfigFlags(config *rest.Config, namespace string) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		Namespace:   &namespace,
		APIServer:   &config.Host,
		CAFile:      &config.CAFile,
		BearerToken: &config.BearerToken,
	}
}

func newStorageDriver(client *kubernetes.Clientset, logFunc infoLogFunc, namespace, d string) (*storage.Storage, error) {
	switch d {
	case "secret", "secrets", "":
		s := driver.NewSecrets(client.CoreV1().Secrets(namespace))
		s.Log = logFunc
		return storage.Init(s), nil
	case "configmap", "configmaps":
		c := driver.NewConfigMaps(client.CoreV1().ConfigMaps(namespace))
		c.Log = logFunc
		return storage.Init(c), nil
	case "memory":
		m := driver.NewMemory()
		return storage.Init(m), nil
	default:
		return nil, fmt.Errorf("unsupported storage driver '%s'", d)
	}
}
