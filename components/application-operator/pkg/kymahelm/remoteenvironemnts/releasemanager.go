package remoteenvironemnts

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	reChartDirectory = "remote-environments"
)

type ReleaseManager interface {
	InstallNewREChart(name string) (hapi_4.Status_Code, string, error)
	DeleteREChart(name string) error
	CheckReleaseExistence(name string) (bool, error)
	CheckReleaseStatus(name string) (hapi_4.Status_Code, string, error)
}

type releaseManager struct {
	helmClient kymahelm.HelmClient
	overrides  string
	namespace  string
}

func NewReleaseManager(helmClient kymahelm.HelmClient, overrides string, namespace string) ReleaseManager {
	return &releaseManager{
		helmClient: helmClient,
		overrides:  overrides,
		namespace:  namespace,
	}
}

func (r *releaseManager) InstallNewREChart(name string) (hapi_4.Status_Code, string, error) {
	installResponse, err := r.helmClient.InstallReleaseFromChart(reChartDirectory, r.namespace, name, r.overrides)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	return installResponse.Release.Info.Status.Code, installResponse.Release.Info.Description, nil
}

func (r *releaseManager) DeleteREChart(name string) error {
	_, err := r.helmClient.DeleteRelease(name)
	if err != nil {
		return err
	}

	return nil
}

func (r *releaseManager) CheckReleaseExistence(name string) (bool, error) {
	listResponse, err := r.helmClient.ListReleases()
	if err != nil {
		return false, err
	}

	releases := listResponse.Releases

	for _, rel := range releases {
		if rel.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (r *releaseManager) CheckReleaseStatus(name string) (hapi_4.Status_Code, string, error) {
	status, err := r.helmClient.ReleaseStatus(name)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	return status.Info.Status.Code, status.Info.Description, nil
}
