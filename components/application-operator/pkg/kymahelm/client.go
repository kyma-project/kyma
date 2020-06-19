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
	InstallReleaseFromChart(chartDir, releaseName string, namespace string, overrides map[string]interface{}) (*release.Release, error)
	UpdateReleaseFromChart(chartDir, releaseName string, namespace string, overrides map[string]interface{}) (*release.Release, error)
	DeleteRelease(releaseName string, namespace string) (*release.UninstallReleaseResponse, error)
	ReleaseStatus(releaseName string, namespace string) (*release.Release, error)
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
	installAction.Timeout = time.Duration(hc.installationTimeout) * time.Second
	installAction.Wait = true

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
