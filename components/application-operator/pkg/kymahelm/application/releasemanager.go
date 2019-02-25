package application

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm"
	"github.com/pkg/errors"
	hapi_4 "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	applicationChartDirectory = "application"
)

type ReleaseManager interface {
	InstallChart(application *v1alpha1.Application) (hapi_4.Status_Code, string, error)
	DeleteReleaseIfExists(name string) error
	CheckReleaseExistence(name string) (bool, error)
	CheckReleaseStatus(name string) (hapi_4.Status_Code, string, error)
}

type releaseManager struct {
	helmClient        kymahelm.HelmClient
	overridesDefaults OverridesData
	namespace         string
}

func NewReleaseManager(helmClient kymahelm.HelmClient, overridesDefaults OverridesData, namespace string) ReleaseManager {
	return &releaseManager{
		helmClient:        helmClient,
		overridesDefaults: overridesDefaults,
		namespace:         namespace,
	}
}

func (r *releaseManager) InstallChart(application *v1alpha1.Application) (hapi_4.Status_Code, string, error) {
	overrides, err := r.prepareOverrides(application)
	if err != nil {
		return hapi_4.Status_FAILED, "", errors.Wrapf(err, "Error parsing overrides for %s Application", application.Name)
	}

	installResponse, err := r.helmClient.InstallReleaseFromChart(applicationChartDirectory, r.namespace, application.Name, overrides)
	if err != nil {
		return hapi_4.Status_FAILED, "", err
	}

	return installResponse.Release.Info.Status.Code, installResponse.Release.Info.Description, nil
}

func (r *releaseManager) prepareOverrides(application *v1alpha1.Application) (string, error) {
	overridesData := r.overridesDefaults
	if application.Spec.HasTenant() == true && application.Spec.HasGroup() == true {
		overridesData.SubjectCN = fmt.Sprintf("%s\\\\\\;%s\\\\\\;%s", application.Spec.Tenant, application.Spec.Group, application.Name)
	} else {
		overridesData.SubjectCN = application.Name
	}

	return kymahelm.ParseOverrides(overridesData, overridesTemplate)
}

func (r *releaseManager) DeleteReleaseIfExists(name string) error {
	releaseExist, err := r.checkExistence(name)
	if err != nil {
		return err
	}

	if releaseExist {
		_, err := r.helmClient.DeleteRelease(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *releaseManager) CheckReleaseExistence(name string) (bool, error) {
	return r.checkExistence(name)
}

func (r *releaseManager) checkExistence(name string) (bool, error) {
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
