package kymahelm

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/klog"
	"os"
	"time"
)

//go:generate mockery -name HelmClient
type HelmClient interface {
	ListReleases(namespace string) ([]*release.Release, error)
	InstallReleaseFromChart(chartDir, ns, releaseName, overrides string) (*release.Release, error)
	UpdateReleaseFromChart(chartDir, releaseName, overrides string) (*release.Release, error)
	DeleteRelease(releaseName string) (*release.UninstallReleaseResponse, error)
	ReleaseStatus(rlsName string) (*release.Release, error)
}

type helmClient struct {
	namespace           string
	installationTimeout int64
	settings            *cli.EnvSettings
}

func NewClient(namespace string, host, tlsKeyFile, tlsCertFile string, skipVerify bool, installationTimeout int64) (HelmClient, error) {

	return &helmClient{
		//helm:                helm.NewClient(helm.Host(host), helm.WithTLS(tlsCfg)),
		settings:            cli.New(),
		namespace:           namespace,
		installationTimeout: installationTimeout,
	}, nil
}

func (hc *helmClient) ListReleases(namespace string) ([]*release.Release, error) {

	actionConfig, err := hc.actionConfigInit()
	if err != nil {
		return nil, err
	}

	client := action.NewList(&actionConfig)
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

func (hc *helmClient) InstallReleaseFromChart(chartDir, ns, releaseName, overrides string) (*release.Release, error) {

	actionConfig, err := hc.actionConfigInit()
	if err != nil {
		return nil, err
	}
	client := action.NewInstall(&actionConfig)

	client.ReleaseName = releaseName
	client.Timeout = time.Duration(hc.installationTimeout)
	client.Namespace = hc.namespace

	fullPath, err := client.ChartPathOptions.LocateChart(chartDir, hc.settings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(fullPath)
	if err != nil {
		return nil, err
	}

	//if req := chartRequested.Metadata.Dependencies; req != nil {
	//	if err := action.CheckDependencies(chartRequested, req); err != nil {
	//		return nil, err
	//	}
	//}

	valueOpts := hc.getOptions(overrides)
	response, err := client.Run(chartRequested, valueOpts)

	return response, nil

	//return hc.helm.InstallRelease(
	//	chartDir,
	//	ns,
	//	helm.ReleaseName(string(releaseName)),
	//	helm.ValueOverrides([]byte(overrides)), //Without it default "values.yaml" file is ignored!
	//	helm.InstallWait(true),
	//	helm.InstallTimeout(hc.installationTimeout),
	//)
}

// not ready
func (hc *helmClient) UpdateReleaseFromChart(chartDir, releaseName, overrides string) (*release.Release, error) {

	actionConfig, err := hc.actionConfigInit()
	if err != nil {
		return nil, err
	}
	client := action.NewUpgrade(&actionConfig)

	client.Timeout = time.Duration(hc.installationTimeout)
	client.Namespace = hc.namespace

	fullPath, err := client.ChartPathOptions.LocateChart(chartDir, hc.settings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(fullPath)
	if err != nil {
		return nil, err
	}

	//if req := chartRequested.Metadata.Dependencies; req != nil {
	//	if err := action.CheckDependencies(chartRequested, req); err != nil {
	//		return nil, err
	//	}
	//}

	valueOpts := hc.getOptions(overrides)
	response, err := client.Run(releaseName, chartRequested, valueOpts)

	return response, nil


	//return hc.helm.UpdateRelease(
	//	releaseName,
	//	chartDir,
	//	helm.UpgradeTimeout(hc.installationTimeout),
	//	helm.UpdateValueOverrides([]byte(overrides)),
	//)
}

func (hc *helmClient) DeleteRelease(releaseName string) (*release.UninstallReleaseResponse, error) {
	//return hc.helm.DeleteRelease(
	//	releaseName,
	//	helm.DeletePurge(true),
	//	helm.DeleteTimeout(hc.installationTimeout),
	//)

	actionConfig, err := hc.actionConfigInit()
	if err != nil {
		return nil, err
	}
	client := action.NewUninstall(&actionConfig)
	response, err := client.Run(releaseName)
	if err != nil {
		return nil, err
	}

	return response, nil
}


func (hc *helmClient) ReleaseStatus(rlsName string) (*release.Release, error) {
	// return hc.helm.ReleaseStatus(rlsName)

	actionConfig, err := hc.actionConfigInit()
	if err != nil {
		return nil , err
	}

	client := action.NewStatus(&actionConfig)
	release, err := client.Run(rlsName)
	if err != nil {
		return nil, err
	}

	return release, nil
}


func (hc *helmClient) actionConfigInit() (action.Configuration, error) {

	actionConfig := new(action.Configuration)
	err := actionConfig.Init(hc.settings.RESTClientGetter(), hc.namespace, os.Getenv("HELM_DRIVER"), klog.Infof)
	if err != nil {
		return *actionConfig, err
	}

	return *actionConfig, nil
}

func (hc *helmClient) getOptions(s string) map[string]interface{} {

	valueOpts := &values.Options{}
	// this will probably fail must find a way to pass overrides
	//
	// 	// values  vals map[string]interface{}
	//	//emptyValues := map[string]interface{}{}
	p := getter.All(hc.settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return vals
	}

	return map[string]interface{}{}
}
