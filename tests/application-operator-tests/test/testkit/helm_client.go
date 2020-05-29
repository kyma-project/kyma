package testkit

/*
Old code

import (
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/tlsutil"

	"time"
)

type HelmClient interface {
	CheckReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error)
	CheckReleaseExistence(name string) (bool, error)
	IsInstalled(rlsName string) bool
	TestRelease(rlsName string) (<-chan *rls.TestReleaseResponse, <-chan error)
}

type helmClient struct {
	helm          *helm.Client
	retryCount    int
	retryWaitTime time.Duration
}

func NewHelmClient(host, tlsKeyFile, tlsCertFile string, skipVerify bool) (HelmClient, error) {
	tlsOpts := tlsutil.Options{
		KeyFile:            tlsKeyFile,
		CertFile:           tlsCertFile,
		InsecureSkipVerify: skipVerify,
	}

	tlsCfg, err := tlsutil.ClientConfig(tlsOpts)
	if err != nil {
		return nil, err
	}

	return &helmClient{
		helm: helm.NewClient(helm.Host(host), helm.WithTLS(tlsCfg)),
	}, nil
}

func (hc *helmClient) CheckReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error) {
	return hc.helm.ReleaseStatus(rlsName)
}

func (hc *helmClient) IsInstalled(rlsName string) bool {
	status, err := hc.CheckReleaseStatus(rlsName)
	return err == nil && status.Info.Status.Code == release.Status_DEPLOYED
}

func (hc *helmClient) CheckReleaseExistence(rlsName string) (bool, error) {
	listResponse, err := hc.helm.ListReleases(helm.ReleaseListStatuses([]release.Status_Code{
		release.Status_DELETED,
		release.Status_DELETING,
		release.Status_DEPLOYED,
		release.Status_FAILED,
		release.Status_PENDING_INSTALL,
		release.Status_PENDING_ROLLBACK,
		release.Status_PENDING_UPGRADE,
		release.Status_SUPERSEDED,
		release.Status_UNKNOWN,
	}))
	if err != nil {
		return false, err
	}

	for _, rel := range listResponse.Releases {
		if rel.Name == rlsName {
			return true, nil
		}
	}
	return false, nil
}

func (hc *helmClient) TestRelease(rlsName string) (<-chan *rls.TestReleaseResponse, <-chan error) {
	return hc.helm.RunReleaseTest(rlsName)
}*/

import (
	"k8s.io/klog"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

type HelmClient interface {
	CheckReleaseStatus(rlsName string, namespace string) (release.Status, error)
	CheckReleaseExistence(rlsName string, namespace string) (bool, error)
	IsInstalled(rlsName string, namespace string) bool
	TestRelease(rlsName string, namespace string) (*release.Release, error)
}

type helmClient struct {
	retryCount    int
	retryWaitTime time.Duration
	helmDriver    string
}

func NewHelmClient(helmDriver string) (HelmClient, error) {

	return &helmClient{
		helmDriver: helmDriver,
	}, nil
}

func (hc *helmClient) CheckReleaseStatus(rlsName string, namespace string) (release.Status, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return release.StatusUnknown, err
	}

	statusAction := action.NewStatus(actionConfig)
	status, err := statusAction.Run(rlsName)
	if err != nil {
		return release.StatusUnknown, err
	}

	return status.Info.Status, nil
}

func (hc *helmClient) IsInstalled(rlsName string, namespace string) bool {

	status, err := hc.CheckReleaseStatus(rlsName, namespace)
	return err == nil && status == release.StatusDeployed
}

func (hc *helmClient) listReleases(namespace string) ([]*release.Release, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	listAction := action.NewList(actionConfig)

	listAction.Deployed = true
	listAction.Uninstalled = true
	listAction.Superseded = true
	listAction.Uninstalling = true
	listAction.Deployed = true
	listAction.Failed = true
	listAction.Pending = true

	results, err := listAction.Run()
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (hc *helmClient) TestRelease(rlsName string, namespace string) (*release.Release, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	releaseTesting := action.NewReleaseTesting(actionConfig)
	releaseTesting.Namespace = namespace

	return releaseTesting.Run(rlsName)
	/*
		// We only return an error if we weren't even able to get the
					// release, otherwise we keep going so we can print status and logs
					// if requested
					if runErr != nil && rel == nil {
						return runErr
					}

					if err := outfmt.Write(out, &statusPrinter{rel, settings.Debug}); err != nil {
						return err
					}

					if outputLogs {
						// Print a newline to stdout to separate the output
						fmt.Fprintln(out)
						if err := client.GetPodLogs(out, rel); err != nil {
							return err
						}
					}
	*/

	//if err != nil {
	//	for _, hook := range res.Hooks {
	//		lastRun := hook.LastRun
	//
	//		if lastRun.Phase == release.HookPhaseFailed {
	//			return microerror.Maskf(testReleaseFailureError, "tests for %#q failed", releaseName)
	//		}
	//	}
	//}
}

func (hc *helmClient) CheckReleaseExistence(rlsName string, namespace string) (bool, error) {

	releases, err := hc.listReleases(namespace)

	if err != nil {
		return false, err
	}

	for _, rel := range releases {
		if rel.Name == rlsName {
			return true, nil
		}
	}

	return false, nil
}

func (hc *helmClient) actionConfigInit(namespace string) (*action.Configuration, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	kubeConfig := genericclioptions.NewConfigFlags(false)
	kubeConfig.APIServer = &config.Host
	kubeConfig.BearerToken = &config.BearerToken
	kubeConfig.CAFile = &config.CAFile
	kubeConfig.Namespace = &namespace

	actionConfig := new(action.Configuration)

	err = actionConfig.Init(kubeConfig, namespace, hc.helmDriver, klog.Infof)
	if err != nil {
		return actionConfig, err
	}

	return actionConfig, nil
}

/*
package kymahelm

import (
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

//go:generate mockery -name HelmClient
type HelmClient interface {
	ListReleases(namespace string) ([]*release.Release, error)
	ReleaseStatus(releaseName string, namespace string) (*release.Release, error)


  	CheckReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error)
	CheckReleaseExistence(name string) (bool, error)
	IsInstalled(rlsName string) bool
	TestRelease(rlsName string) (<-chan *rls.TestReleaseResponse, <-chan error)

}

type helmClient struct {
	installationTimeout int64
	settings            *cli.EnvSettings
	config              *rest.Config
	helmdriver          string
}

func NewClient(config *rest.Config, driver string, installationTimeout int64) (HelmClient, error) {

	return &helmClient{
		config:              config,
		helmdriver:          driver,
		settings:            cli.New(),
		installationTimeout: installationTimeout,
	}, nil
}

func (hc *helmClient) ListReleases(namespace string) ([]*release.Release, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	listAction := action.NewList(actionConfig)

	listAction.Deployed = true
	listAction.Uninstalled = true
	listAction.Superseded = true
	listAction.Uninstalling = true
	listAction.Deployed = true
	listAction.Failed = true
	listAction.Pending = true

	results, err := listAction.Run()
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (hc *helmClient) InstallReleaseFromChart(chartDir, releaseName string, namespace string, overrides map[string]interface{}) (*release.Release, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	installAction := action.NewInstall(actionConfig)

	installAction.ReleaseName = releaseName
	installAction.Namespace = namespace
	installAction.Timeout = time.Duration(hc.installationTimeout)

	chartRequested, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	response, err := installAction.Run(chartRequested, overrides)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (hc *helmClient) UpdateReleaseFromChart(chartDir, releaseName string, namespace string, overrides map[string]interface{}) (*release.Release, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	upgradeAction := action.NewUpgrade(actionConfig)

	upgradeAction.Namespace = namespace
	upgradeAction.Timeout = time.Duration(hc.installationTimeout)

	chartRequested, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	response, err := upgradeAction.Run(releaseName, chartRequested, overrides)
	if err != nil {
		return nil, err
	}

	return response, err
}

func (hc *helmClient) DeleteRelease(releaseName string, namespace string) (*release.UninstallReleaseResponse, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	uninstallAction := action.NewUninstall(actionConfig)
	uninstallAction.Timeout = time.Duration(hc.installationTimeout)

	response, err := uninstallAction.Run(releaseName)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (hc *helmClient) ReleaseStatus(releaseName string, namespace string) (*release.Release, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	statusAction := action.NewStatus(actionConfig)
	release, err := statusAction.Run(releaseName)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func (hc *helmClient) actionConfigInit(namespace string) (*action.Configuration, error) {

	config := hc.config

	kubeConfig := genericclioptions.NewConfigFlags(false)
	kubeConfig.APIServer = &config.Host
	kubeConfig.BearerToken = &config.BearerToken
	kubeConfig.CAFile = &config.CAFile

	actionConfig := new(action.Configuration)
	err := actionConfig.Init(kubeConfig, namespace, hc.helmdriver, klog.Infof)
	if err != nil {
		return actionConfig, err
	}

	return actionConfig, nil
}



*/
