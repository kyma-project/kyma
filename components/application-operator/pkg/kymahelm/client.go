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

	client := action.NewList(actionConfig)

	client.Deployed = true
	client.Uninstalled = true
	client.Superseded = true
	client.Uninstalling = true
	client.Deployed = true
	client.Failed = true
	client.Pending = true

	results, err := client.Run()
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
	client := action.NewInstall(actionConfig)

	client.ReleaseName = releaseName
	client.Namespace = namespace
	client.Timeout = time.Duration(hc.installationTimeout)

	chartRequested, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	response, err := client.Run(chartRequested, overrides)
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
	client := action.NewUpgrade(actionConfig)

	client.Namespace = namespace
	client.Timeout = time.Duration(hc.installationTimeout)

	chartRequested, err := loader.Load(chartDir)
	if err != nil {
		return nil, err
	}

	response, err := client.Run(releaseName, chartRequested, overrides)

	return response, nil
}

func (hc *helmClient) DeleteRelease(releaseName string, namespace string) (*release.UninstallReleaseResponse, error) {

	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	client := action.NewUninstall(actionConfig)
	client.Timeout = time.Duration(hc.installationTimeout)

	response, err := client.Run(releaseName)
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

	client := action.NewStatus(actionConfig)
	release, err := client.Run(releaseName)
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
