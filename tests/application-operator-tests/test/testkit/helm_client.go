package testkit

import (
	"time"

	log "github.com/sirupsen/logrus"
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
	GetRelease(rlsName string, namespace string) (*release.Release, error)
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

	if err != nil {
		return false
	}
	return status == release.StatusDeployed
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

	err = actionConfig.Init(kubeConfig, namespace, hc.helmDriver, log.Infof)
	if err != nil {
		return actionConfig, err
	}

	return actionConfig, nil
}

func (hc *helmClient) GetRelease(rlsName string, namespace string) (*release.Release, error) {
	actionConfig, err := hc.actionConfigInit(namespace)
	if err != nil {
		return nil, err
	}

	releaseGet := action.NewGet(actionConfig)
	releaseGet.Version = 0

	return releaseGet.Run(rlsName)
}
